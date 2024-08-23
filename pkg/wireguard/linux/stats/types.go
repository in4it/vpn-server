//go:build linux
// +build linux

package stats

import "time"

type PeerStat struct {
	Timestamp         time.Time `json:"timestamp"`
	PublicKey         string    `json:"publicKey"`
	LastHandshakeTime time.Time `json:"lastHandshakeTime"`
	ReceiveBytes      int64     `json:"receiveBytes"`
	TransmitBytes     int64     `json:"transmitBytes"`
}
