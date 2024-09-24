package observability

import (
	"bufio"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func (o *Observability) getLogs(fromDate, endDate time.Time, pos int64, maxLogLines, offset int, search string) (LogEntryResponse, error) {
	logEntryResponse := LogEntryResponse{
		Enabled:      true,
		Environments: []string{"dev", "qa", "prod"},
		LogEntries:   []LogEntry{},
		Keys:         KeyValueInt{},
	}

	keys := make(map[KeyValue]int)

	logFiles := []string{}

	if maxLogLines == 0 {
		maxLogLines = 100
	}

	for d := fromDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		fileList, err := o.Storage.ReadDir(d.Format("2006/01/02"))
		if err != nil {
			logEntryResponse.NextPos = -1
			return logEntryResponse, nil // can't read directory, return empty response
		}
		for _, filename := range fileList {
			logFiles = append(logFiles, d.Format("2006/01/02")+"/"+filename)
		}
	}

	fileReaders, err := o.Storage.OpenFilesFromPos(logFiles, pos)
	if err != nil {
		return logEntryResponse, fmt.Errorf("error while reading files: %s", err)
	}
	for _, fileReader := range fileReaders {
		defer fileReader.Close()
	}

	for _, logInputData := range fileReaders { // read multiple files
		if len(logEntryResponse.LogEntries) >= maxLogLines {
			break
		}
		scanner := bufio.NewScanner(logInputData)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = scanMessage(data, atEOF)
			pos += int64(advance)
			return
		})
		for scanner.Scan() && len(logEntryResponse.LogEntries) < maxLogLines { // read multiple lines
			// decode, store as logentry
			logMessage := decodeMessage(scanner.Bytes())
			logline, ok := logMessage.Data["log"]
			if ok {
				timestamp := floatToDate(logMessage.Date).Add(time.Duration(offset) * time.Minute)
				if search == "" || strings.Contains(logline, search) {
					logEntry := LogEntry{
						Timestamp: timestamp.Format(TIMESTAMP_FORMAT),
						Data:      logline,
					}
					logEntryResponse.LogEntries = append(logEntryResponse.LogEntries, logEntry)
					for k, v := range logMessage.Data {
						if k != "log" {
							keys[KeyValue{Key: k, Value: v}] += 1
						}
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return logEntryResponse, fmt.Errorf("log file read (scanner) error: %s", err)
		}
	}
	if len(logEntryResponse.LogEntries) < maxLogLines {
		logEntryResponse.NextPos = -1 // no more records
	} else {
		logEntryResponse.NextPos = pos
	}

	for k, v := range keys {
		logEntryResponse.Keys = append(logEntryResponse.Keys, KeyValueTotal{
			Key:   k.Key,
			Value: k.Value,
			Total: v,
		})
	}
	sort.Sort(logEntryResponse.Keys)

	return logEntryResponse, nil
}

func floatToDate(datetime float64) time.Time {
	datetimeInt := int64(datetime)
	decimals := datetime - float64(datetimeInt)
	nsecs := int64(math.Round(decimals * 1_000_000)) // precision to match golang's time.Time
	return time.Unix(datetimeInt, nsecs*1000)
}
