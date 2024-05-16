package scim

import (
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
)

func New(storage storage.Iface, userStore *users.UserStore, token string) *scim {
	s := &scim{
		Token:     token,
		UserStore: userStore,
		storage:   storage,
	}
	return s
}
