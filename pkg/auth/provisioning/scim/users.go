package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// handler for multiple users
func (s *scim) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getUsersHandler(w, r)
		return
	case http.MethodPost:
		s.postUsersHandler(w, r)
		return
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

// handler for a single user
func (s *scim) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getUserHandler(w, r)
		return
	case http.MethodPut:
		s.putUserHandler(w, r)
		return
	case http.MethodDelete:
		s.deleteUserHandler(w, r)
		return
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (s *scim) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	attributes := r.URL.Query().Get("attributes")
	filter := r.URL.Query().Get("filter")
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		count = -1
	}
	start, err := strconv.Atoi(r.URL.Query().Get("startIndex"))
	if err != nil {
		start = 1
	}

	if filter != "" {
		response, err := getUsersWithFilter(s.UserStore, attributes, filter)
		if err != nil {
			returnError(w, fmt.Errorf("get user with filter error: %s", err), http.StatusBadRequest)
			return
		}
		write(w, response)
		return
	}
	response, err := getUsersWithoutFilter(s.UserStore, attributes, count, start)
	if err != nil {
		returnError(w, fmt.Errorf("get user with filter error: %s", err), http.StatusBadRequest)
		return
	}
	write(w, response)
}

func (s *scim) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.UserStore.GetUserByID(r.PathValue("id"))
	if err != nil {
		returnError(w, fmt.Errorf("get user by id error: %s", err), http.StatusBadRequest)
		return
	}

	response, err := userResponse(user)
	if err != nil {
		returnError(w, fmt.Errorf("user response error: %s", err), http.StatusBadRequest)
		return
	}

	write(w, response)
}
func (s *scim) putUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.UserStore.GetUserByID(r.PathValue("id"))
	if err != nil {
		returnError(w, fmt.Errorf("get user by id error: %s", err), http.StatusBadRequest)
		return
	}

	var putUserRequest PostUserRequest
	err = json.NewDecoder(r.Body).Decode(&putUserRequest)
	if err != nil {
		returnError(w, fmt.Errorf("unable to decode request payload"), http.StatusBadRequest)
		return
	}

	if !putUserRequest.Active && !user.Suspended { // user is suspended
		err = wireguard.DisableAllClientConfigs(s.storage, user.ID)
		if err != nil {
			returnError(w, fmt.Errorf("could not delete all clients for user %s: %s", user.ID, err), http.StatusBadRequest)
			return
		}
	}
	if putUserRequest.Active && user.Suspended { // user is unsuspended
		err := wireguard.ReactivateAllClientConfigs(s.storage, user.ID)
		if err != nil {
			returnError(w, fmt.Errorf("could not reactivate all clients for user %s: %s", user.ID, err), http.StatusBadRequest)
			return
		}
	}

	user.Suspended = !putUserRequest.Active
	username := getUsername(putUserRequest)
	if user.Login != username {
		if !s.UserStore.LoginExists(username) {
			user.Login = username
		}
	}

	err = s.UserStore.UpdateUser(user)
	if err != nil {
		returnError(w, fmt.Errorf("user update error: %s", err), http.StatusBadRequest)
		return
	}

	response, err := userResponse(user)
	if err != nil {
		returnError(w, fmt.Errorf("user response error: %s", err), http.StatusBadRequest)
		return
	}

	write(w, response)
}

func (s *scim) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.UserStore.GetUserByID(r.PathValue("id"))
	if err != nil {
		returnError(w, fmt.Errorf("get user by id error: %s", err), http.StatusBadRequest)
		return
	}

	err = wireguard.DeleteAllClientConfigs(s.storage, user.ID)
	if err != nil {
		returnError(w, fmt.Errorf("could not delete all clients for user %s: %s", user.ID, err), http.StatusBadRequest)
		return
	}

	err = s.UserStore.DeleteUserByID(user.ID)
	if err != nil {
		returnError(w, fmt.Errorf("user update error: %s", err), http.StatusBadRequest)
		return
	}

	write(w, []byte(""))
}

func (s *scim) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	var postUserRequest PostUserRequest
	err := json.NewDecoder(r.Body).Decode(&postUserRequest)
	if err != nil {
		returnError(w, fmt.Errorf("unable to decode request payload"), http.StatusBadRequest)
		return
	}

	username := getUsername(postUserRequest)

	if s.UserStore.LoginExists(username) {
		writeWithStatus(w, []byte("user already exists"), http.StatusConflict)
		return
	}

	if s.UserStore.GetMaxUsers()-s.UserStore.UserCount() <= 0 {
		writeWithStatus(w, []byte("no license available to add new user"), http.StatusBadRequest)
		return
	}

	user, err := s.UserStore.AddUser(users.User{
		Login:       username,
		Role:        "user",
		Provisioned: true,
		ExternalID:  postUserRequest.ExternalID,
	})
	if err != nil {
		returnError(w, fmt.Errorf("unable to add user: %s", err), http.StatusBadRequest)
		return
	}
	response, err := userResponse(user)
	if err != nil {
		returnError(w, fmt.Errorf("unable to generate user response: %s", err), http.StatusBadRequest)
		return
	}
	writeWithStatus(w, response, http.StatusCreated)
}

func getUsername(postUserRequest PostUserRequest) string {
	username := postUserRequest.UserName
	if username == "" {
		for _, email := range postUserRequest.Emails {
			if email.Primary {
				username = email.Value
			}
		}
	}
	return username
}

func getUsersWithFilter(userStore *users.UserStore, attributes, filter string) ([]byte, error) {
	filterSplit := strings.Split(filter, " ")
	if len(filterSplit) != 3 {
		return []byte{}, fmt.Errorf("invalid filter")
	}
	if strings.ToLower(filterSplit[0]) == "username" {
		if strings.ToLower(filterSplit[1]) == "eq" {
			if userStore.LoginExists(strings.Trim(filterSplit[2], `"`)) {
				user, err := userStore.GetUserByLogin(strings.Trim(filterSplit[2], `"`))
				if err != nil {
					return []byte{}, fmt.Errorf("get user by login error: %s", err)
				}
				response, err := listUserResponse([]users.User{user}, attributes, -1, -1)
				if err != nil {
					return []byte{}, fmt.Errorf("userResponse error: %s", err)
				}
				return response, nil
			}
		}
	}
	response, err := listUserResponse([]users.User{}, attributes, -1, -1)
	if err != nil {
		return response, fmt.Errorf("userResponse error: %s", err)
	}
	return response, nil
}

func getUsersWithoutFilter(userStore *users.UserStore, attributes string, count, start int) ([]byte, error) {
	users := userStore.ListUsers()
	response, err := listUserResponse(users, attributes, count, start)
	if err != nil {
		return []byte{}, fmt.Errorf("userResponse error: %s", err)
	}
	return response, nil
}
