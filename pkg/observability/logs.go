package observability

import (
	"bufio"
	"fmt"
	"math"
	"time"
)

func (o *Observability) getLogs(fromDate, endDate time.Time, pos int64, offset, maxLogLines int) (LogEntryResponse, error) {
	logEntryResponse := LogEntryResponse{}

	logFiles := []string{}

	for d := fromDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		fileList, err := o.Storage.ReadDir(d.Format("2006/01/02"))
		if err != nil {
			return logEntryResponse, fmt.Errorf("can't read log directly: %s", err)
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
			val, ok := logMessage.Data["log"]
			if ok {
				timestamp := floatToDate(logMessage.Date).Add(time.Duration(offset) * time.Minute)
				logEntry := LogEntry{
					Timestamp: timestamp.Format(TIMESTAMP_FORMAT),
					Data:      val,
				}
				logEntryResponse.LogEntries = append(logEntryResponse.LogEntries, logEntry)
			}
		}
		if err := scanner.Err(); err != nil {
			return logEntryResponse, fmt.Errorf("log file read (scanner) error: %s", err)
		}
	}

	return logEntryResponse, nil
}

func floatToDate(datetime float64) time.Time {
	datetimeInt := int64(datetime)
	decimals := datetime - float64(datetimeInt)
	nsecs := int64(math.Round(decimals * 1_000_000)) // precision to match golang's time.Time
	return time.Unix(datetimeInt, nsecs*1000)
}
