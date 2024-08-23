package rest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	var userStatsResponse UserStatsResponse
	statsFile := c.Storage.Client.ConfigPath(path.Join(wireguard.VPN_STATS_DIR, "user-"+time.Now().Format("2006-01-02")) + ".log")
	if !c.Storage.Client.FileExists(statsFile) { // file does not exist so just return empty response
		out, err := json.Marshal(userStatsResponse)
		if err != nil {
			c.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
		return
	}
	logData, err := c.Storage.Client.ReadFile(statsFile)
	if err != nil {
		c.returnError(w, fmt.Errorf("readfile error: %s", err), http.StatusBadRequest)
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(logData))

	receiveBytesLast := make(map[string]int64)
	transmitBytesLast := make(map[string]int64)
	receiveBytesData := make(map[string][]UserStatsDataPoint)
	transmitBytesData := make(map[string][]UserStatsDataPoint)
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
		receiveBytes, err := strconv.ParseInt(inputSplit[3], 10, 64)
		if err == nil {
			if _, ok := receiveBytesData[userID]; !ok {
				receiveBytesData[userID] = []UserStatsDataPoint{}
			}
			receiveBytesData[userID] = append(receiveBytesData[userID], UserStatsDataPoint{X: inputSplit[0], Y: receiveBytes - receiveBytesLast[userID]})
		}
		transmitBytes, err := strconv.ParseInt(inputSplit[4], 10, 64)
		if err == nil {
			if _, ok := transmitBytesData[userID]; !ok {
				transmitBytesData[userID] = []UserStatsDataPoint{}
			}
			transmitBytesData[userID] = append(transmitBytesData[userID], UserStatsDataPoint{X: inputSplit[0], Y: transmitBytes - transmitBytesLast[userID]})
		}
		receiveBytesLast[userID] = receiveBytes
		transmitBytesLast[userID] = transmitBytes
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
	for userID, data := range receiveBytesData {
		userStatsResponse.ReceiveBytes.Datasets = append(userStatsResponse.ReceiveBytes.Datasets, UserStatsDataset{
			Label: userID,
			Data:  data,
		})
	}
	for userID, data := range transmitBytesData {
		userStatsResponse.TransmitBytes.Datasets = append(userStatsResponse.TransmitBytes.Datasets, UserStatsDataset{
			Label: userID,
			Data:  data,
		})
	}

	out, err := json.Marshal(userStatsResponse)
	if err != nil {
		c.returnError(w, fmt.Errorf("user stats response marshal error: %s", err), http.StatusBadRequest)
		return
	}
	c.write(w, out)
}
