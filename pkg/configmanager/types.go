package configmanager

import (
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

type ConfigManager struct {
	PrivateKey  string
	PublicKey   string
	Storage     storage.Iface
	ClientCache *wireguard.ClientCache
	VPNConfig   *wireguard.VPNConfig
}

type UpgradeResponse struct {
	NewVersionAvailable bool   `json:"newVersionAvailable"`
	NewVersion          string `json:"newVersion"`
	CurrentVersion      string `json:"currentVersion"`
}
