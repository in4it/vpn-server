package users

import (
	"encoding/json"
	"fmt"
	"os/user"
	"sync"
)

var UserStoreMu sync.Mutex

func (u *UserStore) SaveUsers() error {
	UserStoreMu.Lock()
	defer UserStoreMu.Unlock()
	out, err := json.Marshal(u.Users)
	if err != nil {
		return fmt.Errorf("user store marshal error: %s", err)
	}
	err = u.storage.WriteFile(u.storage.ConfigPath(USERSTORE_FILENAME), out)
	if err != nil {
		return fmt.Errorf("user store write error: %s", err)
	}
	// fix permissions
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Username != "vpn" {
		err = u.storage.EnsureOwnership(u.storage.ConfigPath(USERSTORE_FILENAME), "vpn")
		if err != nil {
			return fmt.Errorf("ensure ownership error (userstore): %s", err)
		}
	}

	return nil
}
