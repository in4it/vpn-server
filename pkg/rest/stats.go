package rest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("date") == "" {
		c.returnError(w, fmt.Errorf("no date supplied"), http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", r.PathValue("date"))
	if err != nil {
		c.returnError(w, fmt.Errorf("invalid date: %s", err), http.StatusBadRequest)
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
	users := c.UserStore.ListUsers()
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
	if !dateEqual(time.Now(), date) {
		statsFiles = append(statsFiles, path.Join(wireguard.VPN_STATS_DIR, "user-"+date.AddDate(0, 0, 1).Format("2006-01-02")+".log"))
	}
	logData := bytes.NewBuffer([]byte{})
	for _, statsFile := range statsFiles {
		if c.Storage.Client.FileExists(statsFile) {
			fileLogData, err := c.Storage.Client.ReadFile(statsFile)
			if err != nil {
				c.returnError(w, fmt.Errorf("readfile error: %s", err), http.StatusBadRequest)
				return
			}
			logData.Write(fileLogData)
		}
	}

	scanner := bufio.NewScanner(logData)

	receiveBytesLast := make(map[string]int64)
	transmitBytesLast := make(map[string]int64)
	receiveBytesData := make(map[string][]UserStatsDataPoint)
	transmitBytesData := make(map[string][]UserStatsDataPoint)
	handshakeLast := make(map[string]time.Time)
	handshakeData := make(map[string][]UserStatsDataPoint)
	for scanner.Scan() { // all other entries
		inputSplit := strings.Split(scanner.Text(), ",")
		userID := inputSplit[1]
		if _, ok := receiveBytesLast[userID]; !ok {
			val, err := strconv.ParseInt(inputSplit[3], 10, 64)
			if err == nil {
				receiveBytesLast[userID] = val
			} else {
				receiveBytesLast[userID] = 0
			}
		}
		if _, ok := transmitBytesLast[userID]; !ok {
			val, err := strconv.ParseInt(inputSplit[4], 10, 64)
			if err == nil {
				transmitBytesLast[userID] = val
			} else {
				transmitBytesLast[userID] = 0
			}
		}
		if _, ok := handshakeLast[userID]; !ok {
			handshakeLast[userID] = time.Time{}
		}
		receiveBytes, err := strconv.ParseInt(inputSplit[3], 10, 64)
		if err == nil {
			if _, ok := receiveBytesData[userID]; !ok {
				receiveBytesData[userID] = []UserStatsDataPoint{}
			}
			value := math.Round(float64((receiveBytes-receiveBytesLast[userID])/unitAdjustment*100)) / 100
			timestamp, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[0])
			if err == nil {
				timestamp = timestamp.Add(time.Duration(offset) * time.Minute)
				if dateEqual(timestamp, date) {
					receiveBytesData[userID] = append(receiveBytesData[userID], UserStatsDataPoint{X: timestamp.Format(wireguard.TIMESTAMP_FORMAT), Y: value})
				}
			}
		}
		transmitBytes, err := strconv.ParseInt(inputSplit[4], 10, 64)
		if err == nil {
			if _, ok := transmitBytesData[userID]; !ok {
				transmitBytesData[userID] = []UserStatsDataPoint{}
			}
			value := math.Round(float64((transmitBytes-transmitBytesLast[userID])/unitAdjustment*100)) / 100
			timestamp, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[0])
			if err == nil {
				timestamp = timestamp.Add(time.Duration(offset) * time.Minute)
				if dateEqual(timestamp, date) {
					transmitBytesData[userID] = append(transmitBytesData[userID], UserStatsDataPoint{X: timestamp.Format(wireguard.TIMESTAMP_FORMAT), Y: value})
				}
			}
		}
		handshake, err := time.Parse(wireguard.TIMESTAMP_FORMAT, inputSplit[5])
		if err == nil {
			handshake = handshake.Add(time.Duration(offset) * time.Minute)
			if dateEqual(handshake, date) && !handshake.Equal(handshakeLast[userID]) {
				if _, ok := handshakeData[userID]; !ok {
					handshakeData[userID] = []UserStatsDataPoint{}
				}
				handshakeData[userID] = append(handshakeData[userID], UserStatsDataPoint{X: handshake.Format(wireguard.TIMESTAMP_FORMAT), Y: 1})
			}
		}
		receiveBytesLast[userID] = receiveBytes
		transmitBytesLast[userID] = transmitBytes
		handshakeLast[userID] = handshake
	}

	if err := scanner.Err(); err != nil {
		c.returnError(w, fmt.Errorf("log file read (scanner) error: %s", err), http.StatusBadRequest)
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
		c.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
		return
	}
	c.write(w, out)
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

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
