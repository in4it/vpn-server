package commands

import (
	"fmt"

	"github.com/in4it/go-devops-platform/rest"
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/go-devops-platform/users"
)

func ResetAdminMFA(storage storage.Iface) error {
	c, err := rest.GetConfig(storage)
	if err != nil {
		return fmt.Errorf("config retrieval error: %s", err)
	}
	c.UserStore, err = users.NewUserStore(storage, -1)
	if err != nil {
		return fmt.Errorf("userstore initialization error: %s", err)
	}
	if !c.UserStore.LoginExists("admin") {
		return fmt.Errorf("admin user doesn't exist")
	}
	user, err := c.UserStore.GetUserByLogin("admin")
	if err != nil {
		return fmt.Errorf("GetUserByLogin error: %s", err)
	}
	user.Factors = []users.Factor{}

	err = c.UserStore.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("UpdateUser error: %s", err)
	}
	return nil
}
