package scim

import (
	"encoding/json"
	"fmt"

	"github.com/in4it/wireguard-server/pkg/users"
)

func listUserResponse(users []users.User, attributes string, count, start int) ([]byte, error) {
	if start != -1 && start > 1 && start <= len(users) {
		users = users[start:]
	}
	totalResults := len(users)
	if len(users) > count && count != -1 {
		users = users[0:count]
	}
	response := UserResponse{
		TotalResults: totalResults,
		ItemsPerPage: len(users),
		StartIndex:   start,
		Schemas:      getSchemas("ListResponse"),
		Resources:    make([]UserResource, len(users)),
	}
	for k := range users {
		response.Resources[k] = UserResource{
			ID:       users[k].ID,
			UserName: users[k].Login,
		}
	}
	out, err := json.Marshal(response)
	if err != nil {
		return out, fmt.Errorf("json marshal error: %s", err)
	}
	return out, nil
}

func userResponse(user users.User) ([]byte, error) {
	response := PostUserRequest{
		Schemas:  getSchemas("User"),
		Id:       user.ID,
		UserName: user.Login,
		Active:   !user.Suspended,
	}
	out, err := json.Marshal(response)
	if err != nil {
		return out, fmt.Errorf("json marshal error: %s", err)
	}
	return out, nil
}

func getSchemas(responseType string) []string {
	if responseType == "User" {
		return []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	}
	return []string{"urn:ietf:params:scim:api:messages:2.0:" + responseType}

}
