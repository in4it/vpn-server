package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) GetUserFromRequest(r *http.Request) (users.User, error) {
	claims := r.Context().Value(CustomValue("claims")).(jwt.MapClaims)
	sub, ok := claims["sub"]
	if !ok {
		return users.User{}, fmt.Errorf("userinfoHandler: subject not found in token")
	}
	iss, ok := claims["iss"]
	if !ok {
		return users.User{}, fmt.Errorf("userinfoHandler: issuer not found in token")
	}

	kid, ok := claims["kid"]
	if !ok {
		return users.User{}, fmt.Errorf("userinfoHandler: kid not found in token")
	}

	if kid == c.JWTKeysKID {
		user, err := c.UserStore.GetUserByLogin(sub.(string))
		if err != nil {
			return users.User{}, fmt.Errorf("GetUserByLogin: user not found")
		}
		return user, nil
	} else { // user comes from oidc
		oauth2DataIDs := []string{}
		for _, oauth2Data := range c.OIDCStore.OAuth2Data {
			if oauth2Data.Issuer == iss && oauth2Data.Subject == sub {
				oauth2DataIDs = append(oauth2DataIDs, oauth2Data.ID)
			}
		}
		if len(oauth2DataIDs) == 0 {
			return users.User{}, fmt.Errorf("userinfoHandler: couldn't find user in oidc database")
		}
		user, err := c.UserStore.GetUserByOIDCIDs(oauth2DataIDs)
		if err != nil {
			return user, fmt.Errorf("get user by oidc id failed: %s", err)
		}
		return user, nil
	}
}

func (c *Context) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users := c.UserStore.ListUsers()
		userResponse := make([]UsersResponse, len(users))
		for k, user := range users {
			userResponse[k].ID = user.ID
			userResponse[k].Login = user.Login
			userResponse[k].Role = user.Role
			userResponse[k].OIDCID = user.OIDCID
			userResponse[k].SAMLID = user.SAMLID
			userResponse[k].Suspended = user.Suspended
			userResponse[k].Provisioned = user.Provisioned
			userResponse[k].ConnectionsDisabledOnAuthFailure = user.ConnectionsDisabledOnAuthFailure
			if !user.LastLogin.IsZero() {
				userResponse[k].LastLogin = user.LastLogin.UTC().Format(time.RFC3339)
			}
			for _, oauth2Data := range c.OIDCStore.OAuth2Data {
				if oauth2Data.ID == user.OIDCID {
					userResponse[k].LastTokenRenewal = oauth2Data.LastTokenRenewal
				}
			}
		}
		out, err := json.Marshal(userResponse)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var user users.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		if !isAlphaNumeric(user.Login) {
			c.returnError(w, fmt.Errorf("login not valid"), http.StatusBadRequest)
			return
		}
		if user.Login == "" {
			c.returnError(w, fmt.Errorf("login is empty"), http.StatusBadRequest)
			return
		}
		if user.Password == "" {
			c.returnError(w, fmt.Errorf("password is empty"), http.StatusBadRequest)
			return
		}
		if user.Role != "user" && user.Role != "admin" {
			c.returnError(w, fmt.Errorf("invalid role"), http.StatusBadRequest)
			return
		}
		if c.UserStore.UserCount() >= c.LicenseUserCount {
			c.returnError(w, fmt.Errorf("no more licenses available"), http.StatusBadRequest)
			return
		}

		newUser, err := c.UserStore.AddUser(users.User{Login: user.Login, Password: user.Password, Role: user.Role})
		if err != nil {
			c.returnError(w, fmt.Errorf("add user error: %s", err), http.StatusBadRequest)
			return
		}
		out, err := json.Marshal(newUser)
		if err != nil {
			c.returnError(w, fmt.Errorf("new user marshal error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		userID := r.PathValue("id")
		err := c.UserStore.DeleteUserByID(userID)
		if err != nil {
			c.returnError(w, fmt.Errorf("delete user error: %s", err), http.StatusBadRequest)
			return
		}
		err = wireguard.DeleteAllClientConfigs(c.Storage.Client, userID)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not delete all clients for user %s: %s", userID, err), http.StatusBadRequest)
			return
		}
		c.write(w, []byte(`{"deleted": "`+userID+`"}`))
	case http.MethodPatch:
		dbUser, err := c.UserStore.GetUserByID(r.PathValue("id"))
		if err != nil {
			c.returnError(w, fmt.Errorf("user not found: %s", err), http.StatusBadRequest)
			return
		}
		var user users.User
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&user)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		updateUser := false
		if user.Role != "" && dbUser.Role != user.Role {
			dbUser.Role = user.Role
			updateUser = true
		}
		if dbUser.Suspended != user.Suspended {
			dbUser.Suspended = user.Suspended
			updateUser = true
			if user.Suspended { // user is now suspended
				err := wireguard.DisableAllClientConfigs(c.Storage.Client, user.ID)
				if err != nil {
					c.returnError(w, fmt.Errorf("could not delete all clients for user %s: %s", user.ID, err), http.StatusBadRequest)
					return
				}
			} else { // user is now unsuspended
				err := wireguard.ReactivateAllClientConfigs(c.Storage.Client, user.ID)
				if err != nil {
					c.returnError(w, fmt.Errorf("could not reactivate all clients for user %s: %s", user.ID, err), http.StatusBadRequest)
					return
				}
			}
		}
		if updateUser {
			err = c.UserStore.UpdateUser(dbUser)
			if err != nil {
				c.returnError(w, fmt.Errorf("update user error: %s", err), http.StatusBadRequest)
				return
			}
		}
		if user.Password != "" {
			err = c.UserStore.UpdatePassword(user.ID, user.Password)
			if err != nil {
				c.returnError(w, fmt.Errorf("update password error: %s", err), http.StatusBadRequest)
				return
			}
		}
		out, err := json.Marshal(dbUser)
		if err != nil {
			c.returnError(w, fmt.Errorf("marshal dbuser error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func addOrModifyExternalUser(storage storage.Iface, userStore *users.UserStore, login, authType, externalAuthID string) (users.User, error) {
	if userStore.LoginExists(login) {
		existingUser, err := userStore.GetUserByLogin(login)
		if err != nil {
			return existingUser, fmt.Errorf("couldn't find existing user in database: %s", login)
		}

		if authType == "oidc" {
			existingUser.OIDCID = externalAuthID
		}
		if authType == "saml" {
			existingUser.SAMLID = externalAuthID
		}

		if existingUser.ConnectionsDisabledOnAuthFailure { // we can enable connections again after auth
			err := wireguard.ReactivateAllClientConfigs(storage, existingUser.ID)
			if err != nil {
				return existingUser, fmt.Errorf("could not reactivate all clients for user %s: %s", existingUser.ID, err)
			}
			existingUser.ConnectionsDisabledOnAuthFailure = false
		}

		existingUser.LastLogin = time.Now()

		err = userStore.UpdateUser(existingUser)
		if err != nil {
			return existingUser, fmt.Errorf("couldn't update user: %s", login)
		}
		return existingUser, nil
	} else {
		newUser := users.User{
			Login: login,
			Role:  "user",
		}
		if authType == "oidc" {
			newUser.OIDCID = externalAuthID
		}
		if authType == "saml" {
			newUser.SAMLID = externalAuthID
		}

		newUser.LastLogin = time.Now()

		newUserAdded, err := userStore.AddUser(newUser)
		if err != nil {
			return newUserAdded, fmt.Errorf("could not add user: %s", err)
		}
		return newUserAdded, nil
	}
}

func (c *Context) userinfoHandler(w http.ResponseWriter, r *http.Request) {
	var response UserInfoResponse

	user := r.Context().Value(CustomValue("user")).(users.User)

	response.Login = user.Login
	response.Role = user.Role
	if user.OIDCID == "" {
		response.UserType = "local"
	} else {
		response.UserType = "oidc"
	}

	out, err := json.Marshal(response)
	if err != nil {
		c.returnError(w, fmt.Errorf("cannot marshal userinfo response: %s", err), http.StatusBadRequest)
		return
	}
	c.write(w, out)

}
