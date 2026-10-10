package vpn

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/in4it/go-devops-platform/storage"
	dateutils "github.com/in4it/go-devops-platform/utils/date"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

const MAX_LOG_OUTPUT_LINES = 100

func (v *VPN) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("date") == "" {
		v.returnError(w, fmt.Errorf("no date supplied"), http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", r.PathValue("date"))
	if err != nil {
		v.returnError(w, fmt.Errorf("invalid date: %s", err), http.StatusBadRequest)
		return
	}
	unitAdjustment := int64(1)
	switch r.FormValue("unit") {
	case "KB":
		unitAdjustment = 1024
	case "MB":
		unitAdjustment = 1024 * 1024
	case "GB":
		unitAdjustment = 1024 * 1024 * 1024
	}
	offset := 0
	if r.FormValue("offset") != "" {
		i, err := strconv.Atoi(r.FormValue("offset"))
		if err == nil {
			offset = i
		}
	}
	// get all users
	users := v.UserStore.ListUsers()
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.ID] = user.Login
	}
	// calculate stats
	var userStatsResponse UserStatsResponse
	statsFiles := []string{
		path.Join(wireguard.VPN_STATS_DIR, "user-"+date.AddDate(0, 0, -1).Format("2006-01-02")+".log"),
		path.Join(wireguard.VPN_STATS_DIR, "user-"+date.Format("2006-01-02")+".log"),
	}
	if !dateutils.DateEqual(time.Now(), date) {
		statsFiles = append(statsFiles, path.Join(wireguard.VPN_STATS_DIR, "user-"+date.AddDate(0, 0, 1).Format("2006-01-02")+".log"))
	}
	logData := bytes.NewBuffer([]byte{})
	for _, statsFile := range statsFiles {
		if v.Storage.FileExists(statsFile) {
			fileLogData, err := v.Storage.ReadFile(statsFile)
			if err != nil {
				v.returnError(w, fmt.Errorf("readfile error: %s", err), http.StatusBadRequest)
				return
			}
			logData.Write(fileLogData)
		}
	}

	scanner := bufio.NewScanner(logData)

	receiveBytesLast := make(map[string]int64)  // key: userID|connectionID
	transmitBytesLast := make(map[string]int64) // key: userID|connectionID
	receiveBytesData := make(map[string][]UserStatsDataPoint)
	transmitBytesData := make(map[string][]UserStatsDataPoint)
	handshakeLast := make(map[string]time.Time) // key: userID|connectionID
	handshakeData := make(map[string][]UserStatsDataPoint)
	for scanner.Scan() { // all other entries
		inputSplit := strings.Split(scanner.Text(), ",")
		if len(inputSplit) < 6 { // invalid record (blank or truncated line)
			continue
		}
		userID := inputSplit[1]
		connectionID := inputSplit[2]
		key := userID + "|" + connectionID
		if _, ok := receiveBytesLast[key]; !ok {
			val, err := strconv.ParseInt(inputSplit[3], 10, 64)
			if err == nil {
				receiveBytesLast[key] = val
			} else {
				receiveBytesLast[key] = 0
			}
		}
		if _, ok := transmitBytesLast[key]; !ok {
			val, err := strconv.ParseInt(inputSplit[4], 10, 64)
			if err == nil {
				transmitBytesLast[key] = val
			} else {
				transmitBytesLast[key] = 0
			}
		}
		if _, ok := handshakeLast[key]; !ok {
			handshakeLast[key] = time.Time{}
		}
		receiveBytes, err := strconv.ParseInt(inputSplit[3], 10, 64)
		if err == nil {
			if _, ok := receiveBytesData[userID]; !ok {
				receiveBytesData[userID] = []UserStatsDataPoint{}
			}
			value := scaleCounterDelta(receiveBytes, receiveBytesLast[key], unitAdjustment)
			timestamp, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[0])
			if err == nil {
				timestamp = timestamp.Add(time.Duration(offset) * time.Minute)
				if dateutils.DateEqual(timestamp, date) {
					receiveBytesData[userID] = append(receiveBytesData[userID], UserStatsDataPoint{X: timestamp.Format(wireguard.TIMESTAMP_FORMAT), Y: value})
				}
			}
		}
		transmitBytes, err := strconv.ParseInt(inputSplit[4], 10, 64)
		if err == nil {
			if _, ok := transmitBytesData[userID]; !ok {
				transmitBytesData[userID] = []UserStatsDataPoint{}
			}
			value := scaleCounterDelta(transmitBytes, transmitBytesLast[key], unitAdjustment)
			timestamp, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[0])
			if err == nil {
				timestamp = timestamp.Add(time.Duration(offset) * time.Minute)
				if dateutils.DateEqual(timestamp, date) {
					transmitBytesData[userID] = append(transmitBytesData[userID], UserStatsDataPoint{X: timestamp.Format(wireguard.TIMESTAMP_FORMAT), Y: value})
				}
			}
		}
		handshake, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[5])
		if err == nil {
			handshake = handshake.Add(time.Duration(offset) * time.Minute)
			if dateutils.DateEqual(handshake, date) && !handshake.Equal(handshakeLast[key]) {
				if _, ok := handshakeData[userID]; !ok {
					handshakeData[userID] = []UserStatsDataPoint{}
				}
				handshakeData[userID] = append(handshakeData[userID], UserStatsDataPoint{X: handshake.Format(wireguard.TIMESTAMP_FORMAT), Y: 1})
			}
		}
		receiveBytesLast[key] = receiveBytes
		transmitBytesLast[key] = transmitBytes
		handshakeLast[key] = handshake
	}

	if err := scanner.Err(); err != nil {
		v.returnError(w, fmt.Errorf("log file read (scanner) error: %s", err), http.StatusBadRequest)
		return
	}
	userStatsResponse.ReceiveBytes = UserStatsData{
		Datasets: []UserStatsDataset{},
	}
	userStatsResponse.TransmitBytes = UserStatsData{
		Datasets: []UserStatsDataset{},
	}
	userStatsResponse.Handshakes = UserStatsData{
		Datasets: []UserStatsDataset{},
	}
	for userID, data := range receiveBytesData {
		login, ok := userMap[userID]
		if !ok {
			login = "unknown"
		}
		userStatsResponse.ReceiveBytes.Datasets = append(userStatsResponse.ReceiveBytes.Datasets, UserStatsDataset{
			BorderColor:     getColor(len(userStatsResponse.ReceiveBytes.Datasets)),
			BackgroundColor: getColor(len(userStatsResponse.ReceiveBytes.Datasets)),
			Label:           login,
			Data:            data,
			Tension:         0.1,
			ShowLine:        true,
		})
	}
	for userID, data := range transmitBytesData {
		login, ok := userMap[userID]
		if !ok {
			login = "unknown"
		}
		userStatsResponse.TransmitBytes.Datasets = append(userStatsResponse.TransmitBytes.Datasets, UserStatsDataset{
			BorderColor:     getColor(len(userStatsResponse.TransmitBytes.Datasets)),
			BackgroundColor: getColor(len(userStatsResponse.TransmitBytes.Datasets)),
			Label:           login,
			Data:            data,
			Tension:         0.1,
			ShowLine:        true,
		})
	}
	for userID, data := range handshakeData {
		login, ok := userMap[userID]
		if !ok {
			login = "unknown"
		}
		userStatsResponse.Handshakes.Datasets = append(userStatsResponse.Handshakes.Datasets, UserStatsDataset{
			BorderColor:     getColor(len(userStatsResponse.Handshakes.Datasets)),
			BackgroundColor: getColor(len(userStatsResponse.Handshakes.Datasets)),
			Label:           login,
			Data:            data,
			Tension:         0.1,
			ShowLine:        false,
		})
	}

	sort.Sort(userStatsResponse.ReceiveBytes.Datasets)
	sort.Sort(userStatsResponse.TransmitBytes.Datasets)
	sort.Sort(userStatsResponse.Handshakes.Datasets)

	out, err := json.Marshal(userStatsResponse)
	if err != nil {
		v.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
		return
	}
	v.write(w, out)
}

func (v *VPN) packetLogsHandler(w http.ResponseWriter, r *http.Request) {
	vpnConfig, err := wireguard.GetVPNConfig(v.Storage)
	if err != nil {
		v.returnError(w, fmt.Errorf("get vpn config error: %s", err), http.StatusBadRequest)
		return
	}
	if !vpnConfig.EnablePacketLogs { // packet logs is disabled
		out, err := json.Marshal(LogDataResponse{Enabled: false})
		if err != nil {
			v.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
		return
	}
	userID := r.PathValue("user")
	if userID == "" {
		v.returnError(w, fmt.Errorf("no user supplied"), http.StatusBadRequest)
		return
	}
	if !v.isValidPacketLogUserID(userID) {
		v.returnError(w, fmt.Errorf("invalid user supplied"), http.StatusBadRequest)
		return
	}
	if r.PathValue("date") == "" {
		v.returnError(w, fmt.Errorf("no date supplied"), http.StatusBadRequest)
		return
	}
	if !dateRegex.MatchString(r.PathValue("date")) {
		v.returnError(w, fmt.Errorf("invalid date: expected format YYYY-MM-DD"), http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", r.PathValue("date"))
	if err != nil {
		v.returnError(w, fmt.Errorf("invalid date: %s", err), http.StatusBadRequest)
		return
	}
	offset := 0
	if r.FormValue("offset") != "" {
		i, err := strconv.Atoi(r.FormValue("offset"))
		if err == nil {
			offset = i
		}
	}
	pos := int64(0)
	if r.FormValue("pos") != "" {
		i, err := strconv.ParseInt(r.FormValue("pos"), 10, 0)
		if err == nil {
			pos = i
		}
	}
	search := r.FormValue("search")
	// get all users
	users := v.UserStore.ListUsers()
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.ID] = user.Login
	}
	// get filter
	logTypeFilterQueryString := r.URL.Query().Get("logtype")
	logTypeFilter := strings.Split(logTypeFilterQueryString, ",")
	// initialize response
	logData := LogData{
		Schema: LogSchema{
			Columns: map[string]string{
				"Protocol":         "string",
				"Source IP":        "string",
				"Destination IP":   "string",
				"Source Port":      "string",
				"Destination Port": "string",
				"Destination":      "string",
			},
		},
		Data: []LogRow{},
	}
	// logs
	statsFiles := []string{}
	if offset > 0 {
		statsFiles = append(statsFiles, path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, userID+"-"+date.AddDate(0, 0, -1).Format("2006-01-02")+".log"))
	}
	statsFiles = append(statsFiles, path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, userID+"-"+date.Format("2006-01-02")+".log"))
	if !dateutils.DateEqual(time.Now(), date) { // date is in local timezone, and we are UTC, so also read next file
		statsFiles = append(statsFiles, path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, userID+"-"+date.AddDate(0, 0, 1).Format("2006-01-02")+".log"))
	}
	statsFiles, err = getCompressedFilesAndRemoveNonExistent(v.Storage, statsFiles)
	if err != nil {
		v.returnError(w, fmt.Errorf("unable to get files for reading: %s", err), http.StatusBadRequest)
		return
	}
	fileReaders, err := v.Storage.OpenFilesFromPos(statsFiles, pos)
	if err != nil {
		v.returnError(w, fmt.Errorf("error while reading files: %s", err), http.StatusBadRequest)
		return
	}
	for _, fileReader := range fileReaders {
		defer fileReader.Close() //nolint:errcheck
	}

	for _, logInputData := range fileReaders { // read multiple files
		if len(logData.Data) >= MAX_LOG_OUTPUT_LINES {
			break
		}
		scanner := bufio.NewScanner(logInputData)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = bufio.ScanLines(data, atEOF)
			pos += int64(advance)
			return
		})
		for scanner.Scan() && len(logData.Data) < MAX_LOG_OUTPUT_LINES { // read multiple lines
			inputSplit := strings.Split(scanner.Text(), ",")
			if len(inputSplit) < 2 { // invalid record (blank or truncated line)
				continue
			}
			timestamp, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[0])
			if err != nil {
				continue // invalid record
			}
			timestamp = timestamp.Add(time.Duration(offset) * time.Minute)
			if dateutils.DateEqual(timestamp, date) {
				if !filterLogRecord(logTypeFilter, inputSplit[1]) && matchesSearch(search, inputSplit) {
					row := LogRow{
						Timestamp: timestamp.Format("2006-01-02 15:04:05"),
						Data:      inputSplit[1:],
					}
					logData.Data = append(logData.Data, row)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			v.returnError(w, fmt.Errorf("log file read (scanner) error: %s", err), http.StatusBadRequest)
			return
		}
	}
	if len(logData.Data) < MAX_LOG_OUTPUT_LINES {
		pos = -1 // no more records
	}

	// set position
	logData.NextPos = pos

	// logtypes
	packetLogTypes := []string{}
	for k, enabled := range vpnConfig.PacketLogsTypes {
		if enabled {
			packetLogTypes = append(packetLogTypes, k)
		}
	}

	out, err := json.Marshal(LogDataResponse{Enabled: true, LogData: logData, LogTypes: packetLogTypes, Users: userMap})
	if err != nil {
		v.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
		return
	}
	v.write(w, out)
}

var muDecompress sync.Mutex

func getCompressedFilesAndRemoveNonExistent(storage storage.Iface, files []string) ([]string, error) {
	res := []string{}
	for _, file := range files {
		tmpFile := path.Join(wireguard.VPN_PACKETLOGGER_TMP_DIR, path.Base(file))
		if storage.FileExists(file) {
			res = append(res, file)
		} else if storage.FileExists(tmpFile) { // temporary file exists (only created after successful decompression)
			res = append(res, tmpFile)
		} else if storage.FileExists(file + ".gz") { // uncompress log file for random access
			err := decompressToTmpFile(storage, file+".gz", tmpFile)
			if err != nil {
				return res, err
			}
			res = append(res, tmpFile)
		}
	}
	return res, nil
}

// decompressToTmpFile decompresses to a unique temporary file first, and renames it once complete, so that
// concurrent requests never read a partially written file and failed decompressions are never reused
func decompressToTmpFile(storage storage.Iface, compressedFilename, tmpFile string) error {
	muDecompress.Lock()
	defer muDecompress.Unlock()
	if storage.FileExists(tmpFile) { // another request decompressed the file while we were waiting
		return nil
	}
	err := storage.EnsurePath(wireguard.VPN_PACKETLOGGER_TMP_DIR)
	if err != nil {
		return fmt.Errorf("ensure path error: %s", err)
	}
	randomSuffix := make([]byte, 8)
	if _, err := rand.Read(randomSuffix); err != nil {
		return fmt.Errorf("unable to generate tmp file name: %s", err)
	}
	partialFile := tmpFile + ".partial-" + hex.EncodeToString(randomSuffix)
	err = decompressFile(storage, compressedFilename, partialFile)
	if err != nil {
		storage.Remove(partialFile) //nolint:errcheck
		return err
	}
	err = storage.Rename(partialFile, tmpFile)
	if err != nil {
		storage.Remove(partialFile) //nolint:errcheck
		return fmt.Errorf("unable to rename tmp file (%s): %s", tmpFile, err)
	}
	return nil
}

func decompressFile(storage storage.Iface, compressedFilename, outFilename string) error {
	compressedFile, err := storage.OpenFile(compressedFilename)
	if err != nil {
		return fmt.Errorf("unable to open compress filed (%s): %s", compressedFilename, err)
	}
	defer compressedFile.Close() //nolint:errcheck
	gzipReader, err := gzip.NewReader(compressedFile)
	if err != nil {
		return fmt.Errorf("unable to open gzip reader (%s): %s", compressedFilename, err)
	}
	defer gzipReader.Close() //nolint:errcheck
	fileWriter, err := storage.OpenFileForWriting(outFilename)
	if err != nil {
		return fmt.Errorf("unable to open tmp file writer (%s): %s", outFilename, err)
	}
	_, err = io.Copy(fileWriter, gzipReader)
	if err != nil {
		fileWriter.Close() //nolint:errcheck
		return fmt.Errorf("unable to uncompress to file (%s): %s", outFilename, err)
	}
	err = fileWriter.Close()
	if err != nil {
		return fmt.Errorf("unable to close tmp file (%s): %s", outFilename, err)
	}
	return nil
}

// scaleCounterDelta returns the difference between 2 counter values in the requested unit, rounded to 2 decimals.
// When the counter is lower than the previous value, the counter has been reset (e.g. wireguard restart), and the
// current value is the amount since the reset.
func scaleCounterDelta(current, last, unitAdjustment int64) float64 {
	delta := current - last
	if delta < 0 {
		delta = current
	}
	if delta < 0 {
		delta = 0
	}
	return math.Round(float64(delta)/float64(unitAdjustment)*100) / 100
}

var dateRegex = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// isValidPacketLogUserID: the user id is used in a file path, so it must be a uuid or an existing user id.
// "all" is what the webapp requests when no user is selected yet.
func (v *VPN) isValidPacketLogUserID(userID string) bool {
	if userID == "all" {
		return true
	}
	if userID == "" || strings.ContainsAny(userID, `./\`) {
		return false
	}
	if uuidRegex.MatchString(userID) {
		return true
	}
	if v.UserStore != nil {
		if _, err := v.UserStore.GetUserByID(userID); err == nil {
			return true
		}
	}
	return false
}

func getColor(i int) string {
	colors := []string{
		"#DEEFB7",
		"#98DFAF",
		"#5FB49C",
		"#414288",
		"#682D63",
		"#b45f5f",
		"#b49f5f",
		"#8ab45f",
		"#5fb475",
		"#5f8ab4",
		"#755fb4",
		"#b45fb4",
		"#b45f75",
		"#b45f5f",
		"#0066cc",
		"#cc0000",
		"#33cc00",
		"#00cc99",
		"#cc00cc",
		"#00cc99",
	}
	return colors[i%len(colors)]
}

func filterLogRecord(logTypeFilter []string, logType string) bool {
	if len(logTypeFilter) > 0 && logTypeFilter[0] != "" {
		for _, logTypeFilterItem := range logTypeFilter {
			if logType == logTypeFilterItem {
				return false
			}

			if logTypeFilterItem == "dns" && logType == "udp" {
				return false
			}

			splitLogTypes := strings.Split(logTypeFilterItem, "+")
			for _, splitLogType := range splitLogTypes {
				if splitLogType == logType {
					return false
				}
			}
		}
		return true
	}
	return false
}
func matchesSearch(search string, data []string) bool {
	if search == "" {
		return true
	}
	for _, element := range data {
		if strings.Contains(element, search) {
			return true
		}
	}
	return false
}
