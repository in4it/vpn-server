package configmanager

import (
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

type ConfigManager struct {
	PrivateKey  string
	PublicKey   string
	Storage     storage.Iface
	ClientCache *wireguard.ClientCache
}

type UpgradeResponse struct {
	NewVersionAvailable bool   `json:"newVersionAvailable"`
	NewVersion          string `json:"newVersion"`
	CurrentVersion      string `json:"currentVersion"`
}
