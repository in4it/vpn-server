package wireguard

import (
	"net"
	"net/netip"
	"time"
)

type VPNClientData struct {
	Address         string
	DNS             string
	PrivateKey      string
	ServerPublicKey string
	PresharedKey    string
	Endpoint        string
	AllowedIPs      []string
}

type VPNServerData struct {
	Address           string
	PrivateKey        string
	Port              int
	Clients           []VPNServerClient
	DisableNAT        bool
	ExternalInterface string
}

type VPNServerClient struct {
	PublicKey    string
	AllowedIPs   string
	PresharedKey string
}

type VPNConfig struct {
	AddressRange        netip.Prefix    `json:"addressRange"`
	ClientAddressPrefix string          `json:"clientAddressPrefix"`
	PublicKey           string          `json:"publicKey"`
	PresharedKey        string          `json:"presharedKey"`
	Endpoint            string          `json:"endpoint"`
	Port                int             `json:"port"`
	ExternalInterface   string          `json:"externalInterface"`
	Nameservers         []string        `json:"nameservers"`
	DisableNAT          bool            `json:"disableNAT"`
	ClientRoutes        []string        `json:"clientRoutes"`
	EnablePacketLogs    bool            `json:"enablePacketLogs"`
	PacketLogsTypes     map[string]bool `json:"packetLogsTypes"`
	PacketLogsRetention int             `json:"packetLogsRetention"`
}

type PubKeyExchange struct {
	PubKey string `json:"pubKey"`
}

type PeerConfig struct {
	ID               string   `json:"id"`
	DNS              string   `json:"dns"`
	Name             string   `json:"name"`
	ServerAllowedIPs []string `json:"serverAllowedIPs"`
	ClientAllowedIPs []string `json:"clientAllowedIPs"`
	Address          string   `json:"address"`
	PublicKey        string   `json:"publicKey"`
	Disabled         bool     `json:"disabled"`
}
type RefreshClientRequest struct {
	Action    string
	Filenames []string `json:"filenames"`
}

// stats
type StatsEntry struct {
	Timestamp         time.Time
	User              string
	ConnectionID      string
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
}

// client cache

type ClientCache struct {
	Addresses []ClientCacheAddresses
}
type ClientCacheAddresses struct {
	Address  net.IPNet
	ClientID string
}
