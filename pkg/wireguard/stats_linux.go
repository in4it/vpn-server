//go:build linux
// +build linux

package wireguard

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/go-devops-platform/logging"
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard/linux/stats"
)

const RUN_STATS_INTERVAL = 5

func RunStats(storage storage.Iface) {
	err := storage.EnsurePath(VPN_STATS_DIR)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not create stats path: %s. Stats disabled", err))
		return
	}
	err = storage.EnsureOwnership(VPN_STATS_DIR, "vpn")
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not ensure ownership of stats path: %s. Stats disabled", err))
		return
	}
	for {
		err := runStats(storage)
		if err != nil {
			logging.ErrorLog(fmt.Errorf("run stats error: %s", err))
		}
		time.Sleep(RUN_STATS_INTERVAL * time.Minute)
	}
}

func runStats(storage storage.Iface) error {
	peerStats, err := stats.GetStats()
	if err != nil {
		return fmt.Errorf("could not get WireGuard stats: %s", err)
	}

	peerConfigs, err := GetAllPeerConfigs(storage)
	if err != nil {
		return fmt.Errorf("could not get WireGuard peer configs: %s", err)
	}

	statsEntries := []StatsEntry{}

	for _, stat := range peerStats {
		for _, peerConfig := range peerConfigs {
			if stat.PublicKey == peerConfig.PublicKey {
				user, connectionID := splitUserAndConnectionID(peerConfig.ID)
				statsEntries = append(statsEntries, StatsEntry{
					Timestamp:         stat.Timestamp,
					User:              user,
					ConnectionID:      connectionID,
					TransmitBytes:     stat.TransmitBytes,
					ReceiveBytes:      stat.ReceiveBytes,
					LastHandshakeTime: stat.LastHandshakeTime,
				})
			}
		}
	}

	if len(statsEntries) > 0 {
		statsCsv := statsToCsv(statsEntries)

		statsPath := path.Join(VPN_STATS_DIR, "user-"+time.Now().Format("2006-01-02")) + ".log"
		err = storage.AppendFile(statsPath, statsCsv)
		if err != nil {
			return fmt.Errorf("could not append stats to file (%s): %s", statsPath, err)
		}
		err = storage.EnsureOwnership(statsPath, "vpn")
		if err != nil {
			return fmt.Errorf("could not ensure ownership of stats file (%s): %s", statsPath, err)
		}
	}
	return nil
}

func splitUserAndConnectionID(id string) (string, string) {
	split := strings.Split(id, "-")
	if len(split) == 1 {
		return id, ""
	}
	return strings.Join(split[:len(split)-1], "-"), split[len(split)-1]
}

func statsToCsv(statsEntries []StatsEntry) []byte {
	var res bytes.Buffer

	for _, statsEntry := range statsEntries {
		res.WriteString(strings.Join([]string{statsEntry.Timestamp.Format(TIMESTAMP_FORMAT), statsEntry.User, statsEntry.ConnectionID, strconv.FormatInt(statsEntry.ReceiveBytes, 10), strconv.FormatInt(statsEntry.TransmitBytes, 10), statsEntry.LastHandshakeTime.Format(TIMESTAMP_FORMAT)}, ",") + "\n")
	}
	return res.Bytes()
}
