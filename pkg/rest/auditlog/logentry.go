package auditlog

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
)

const TIMESTAMP_FORMAT = "2006-01-02T15:04:05"
const AUDITLOG_STATS_DIR = "stats"

type LogEntry struct {
	Timestamp LogTimestamp `json:"timestamp"`
	UserID    string       `json:"userID"`
	Action    string       `json:"action"`
}
type LogTimestamp time.Time

func (t LogTimestamp) MarshalJSON() ([]byte, error) {
	timestamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(TIMESTAMP_FORMAT))
	return []byte(timestamp), nil
}

func Write(storage storage.Iface, logEntry LogEntry) error {
	statsPath := path.Join(AUDITLOG_STATS_DIR, "logins-"+time.Now().Format("2006-01-02")) + ".log"
	logEntryBytes, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("could not parse log entry: %s", err)
	}
	err = storage.AppendFile(statsPath, logEntryBytes)
	if err != nil {
		return fmt.Errorf("could not append stats to file (%s): %s", statsPath, err)
	}

	return nil
}
