package configmanager

import "github.com/in4it/wireguard-server/pkg/storage"

type ConfigManager struct {
	PrivateKey string
	PublicKey  string
	Storage    storage.Iface
}

type UpgradeResponse struct {
	NewVersionAvailable bool   `json:"newVersionAvailable"`
	NewVersion          string `json:"newVersion"`
	CurrentVersion      string `json:"currentVersion"`
}
