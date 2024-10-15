package vpn

import (
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/go-devops-platform/users"
)

func New(defaultStorage storage.Iface, userStore *users.UserStore) *VPN {
	return &VPN{
		Storage:   defaultStorage,
		UserStore: userStore,
	}
}
