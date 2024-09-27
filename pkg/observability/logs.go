package observability

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (o *Observability) getLogs(fromDate, endDate time.Time, pos int64, maxLogLines, offset int, search string, displayTags []string, filterTags []KeyValue) (LogEntryResponse, error) {
	logEntryResponse := LogEntryResponse{
		Enabled:    true,
		LogEntries: []LogEntry{},
		Tags:       KeyValueInt{},
	}

	keys := make(map[KeyValue]int)

	logFiles := []string{}

	if maxLogLines == 0 {
		maxLogLines = 100
	}

	for d := fromDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		fileList, err := o.Storage.ReadDir(d.Format(DATE_PREFIX))
		if err != nil {
			logEntryResponse.NextPos = -1
			return logEntryResponse, nil // can't read directory, return empty response
		}
		for _, filename := range fileList {
			logFiles = append(logFiles, d.Format(DATE_PREFIX)+"/"+filename)
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
				timestamp := FloatToDate(logMessage.Date).Add(time.Duration(offset) * time.Minute)
				if search == "" || strings.Contains(logline, search) {
					tags := []KeyValue{}
					for _, tag := range displayTags {
						if tagValue, ok := logMessage.Data[tag]; ok {
							tags = append(tags, KeyValue{Key: tag, Value: tagValue})
						}
					}
					filterMessage := true
					if len(filterTags) == 0 {
						filterMessage = false
					} else {
						for _, filter := range filterTags {
							if tagValue, ok := logMessage.Data[filter.Key]; ok {
								if tagValue == filter.Value {
									filterMessage = false
								}
							}
						}
					}
					if !filterMessage {
						logEntry := LogEntry{
							Timestamp: timestamp.Format(TIMESTAMP_FORMAT),
							Data:      logline,
							Tags:      tags,
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
		logEntryResponse.Tags = append(logEntryResponse.Tags, KeyValueTotal{
			Key:   k.Key,
			Value: k.Value,
			Total: v,
		})
	}
	sort.Sort(logEntryResponse.Tags)

	return logEntryResponse, nil
}
