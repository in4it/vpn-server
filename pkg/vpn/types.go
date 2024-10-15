package vpn

import (
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/go-devops-platform/users"
)

type VPN struct {
	Storage   storage.Iface
	UserStore *users.UserStore
	Hostname  string
	Protocol  string
}

type NewConnectionResponse struct {
	Name string `json:"name"`
}
type Connection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserStatsResponse struct {
	ReceiveBytes  UserStatsData `json:"receivedBytes"`
	TransmitBytes UserStatsData `json:"transmitBytes"`
	Handshakes    UserStatsData `json:"handshakes"`
}
type UserStatsData struct {
	Datasets UserStatsDatasets `json:"datasets"`
}
type UserStatsDatasets []UserStatsDataset
type UserStatsDataset struct {
	Label           string               `json:"label"`
	Data            []UserStatsDataPoint `json:"data"`
	Fill            bool                 `json:"fill"`
	BorderColor     string               `json:"borderColor"`
	BackgroundColor string               `json:"backgroundColor"`
	Tension         float64              `json:"tension"`
	ShowLine        bool                 `json:"showLine"`
}

type UserStatsDataPoint struct {
	X string  `json:"x"`
	Y float64 `json:"y"`
}

type LogDataResponse struct {
	LogData  LogData           `json:"logData"`
	Enabled  bool              `json:"enabled"`
	LogTypes []string          `json:"logTypes"`
	Users    map[string]string `json:"users"`
}

type LogData struct {
	Schema  LogSchema `json:"schema"`
	Data    []LogRow  `json:"rows"`
	NextPos int64     `json:"nextPos"`
}
type LogSchema struct {
	Columns map[string]string `json:"columns"`
}
type LogRow struct {
	Timestamp string   `json:"t"`
	Data      []string `json:"d"`
}

type VPNSetupRequest struct {
	Routes              string   `json:"routes"`
	VPNEndpoint         string   `json:"vpnEndpoint"`
	AddressRange        string   `json:"addressRange"`
	ClientAddressPrefix string   `json:"clientAddressPrefix"`
	Port                string   `json:"port"`
	ExternalInterface   string   `json:"externalInterface"`
	Nameservers         string   `json:"nameservers"`
	DisableNAT          bool     `json:"disableNAT"`
	EnablePacketLogs    bool     `json:"enablePacketLogs"`
	PacketLogsTypes     []string `json:"packetLogsTypes"`
	PacketLogsRetention string   `json:"packetLogsRetention"`
}

type TemplateSetupRequest struct {
	ClientTemplate string `json:"clientTemplate"`
	ServerTemplate string `json:"serverTemplate"`
}

type ConnectionLicenseResponse struct {
	LicenseUserCount int `json:"licenseUserCount"`
	ConnectionCount  int `json:"connectionCount"`
}
