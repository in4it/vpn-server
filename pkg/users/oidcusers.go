package users

import "fmt"

func (u *UserStore) GetUserByOIDCIDs(oidcIDs []string) (User, error) {
	for _, user := range u.Users {
		for _, oidcID := range oidcIDs {
			if user.OIDCID == oidcID {
				return user, nil
			}
		}

	}
	return User{}, fmt.Errorf("User not found")
}
