package rest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	// get all users
	users := c.UserStore.ListUsers()
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.ID] = user.Login
	}
	// calculate stats
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
		login, ok := userMap[userID]
		if !ok {
			login = "unknown"
		}
		userStatsResponse.ReceiveBytes.Datasets = append(userStatsResponse.ReceiveBytes.Datasets, UserStatsDataset{
			BorderColor: getColor(len(userStatsResponse.ReceiveBytes.Datasets)),
			Label:       login,
			Data:        data,
			Tension:     0.1,
		})
	}
	for userID, data := range transmitBytesData {
		login, ok := userMap[userID]
		if !ok {
			login = "unknown"
		}
		userStatsResponse.TransmitBytes.Datasets = append(userStatsResponse.TransmitBytes.Datasets, UserStatsDataset{
			BorderColor: getColor(len(userStatsResponse.TransmitBytes.Datasets)),
			Label:       login,
			Data:        data,
			Tension:     0.1,
		})
	}

	sort.Sort(userStatsResponse.ReceiveBytes.Datasets)
	sort.Sort(userStatsResponse.TransmitBytes.Datasets)

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
	}
	return colors[i%len(colors)]
}
