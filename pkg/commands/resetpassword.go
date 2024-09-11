package commands

import (
	"fmt"

	"github.com/in4it/wireguard-server/pkg/rest"
	localstorage "github.com/in4it/wireguard-server/pkg/storage/local"
	"github.com/in4it/wireguard-server/pkg/users"
)

func ResetPassword(appDir, password string) (bool, error) {
	adminCreated := false

	localstorage, err := localstorage.NewWithPath(appDir)
	if err != nil {
		return adminCreated, fmt.Errorf("config retrieval error: %s", err)
	}

	c, err := rest.GetConfig(localstorage)
	if err != nil {
		return adminCreated, fmt.Errorf("config retrieval error: %s", err)
	}
	c.Storage = &rest.Storage{
		Client: localstorage,
	}
	c.UserStore, err = users.NewUserStore(localstorage, -1)
	if err != nil {
		return adminCreated, fmt.Errorf("userstore initialization error: %s", err)
	}
	if c.UserStore.LoginExists("admin") {
		user, err := c.UserStore.GetUserByLogin("admin")
		if err != nil {
			return adminCreated, fmt.Errorf("GetUserByLogin error: %s", err)
		}
		err = c.UserStore.UpdatePassword(user.ID, password)
		if err != nil {
			return adminCreated, fmt.Errorf("UpdatePassword error: %s", err)
		}
		if user.Role != "admin" {
			user.Role = "admin"
			err = c.UserStore.UpdateUser(user)
			if err != nil {
				return adminCreated, fmt.Errorf("UpdateUser error: %s", err)
			}
		}
	} else {
		_, err := c.UserStore.AddUser(users.User{
			Login:    "admin",
			Password: password,
			Role:     "admin",
		})
		adminCreated = true
		if err != nil {
			return adminCreated, fmt.Errorf("admin adduser error: %s", err)
		}
	}
	c.SetupCompleted = true
	err = rest.SaveConfig(c)
	if err != nil {
		return adminCreated, fmt.Errorf("SaveConfig error: %s", err)
	}
	return adminCreated, nil
}
