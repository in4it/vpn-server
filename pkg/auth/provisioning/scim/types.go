package scim

import (
	"net/http"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
)

type scim struct {
	Token     string           `json:"token"`
	UserStore *users.UserStore `json:"userStore"`
	storage   storage.Iface
}

type Iface interface {
	GetRouter() *http.ServeMux
	UpdateToken(token string)
}

type UserResponse struct {
	TotalResults int            `json:"totalResults"`
	ItemsPerPage int            `json:"itemsPerPage"`
	StartIndex   int            `json:"startIndex"`
	Schemas      []string       `json:"schemas"`
	Resources    []UserResource `json:"Resources"`
}
type UserResource struct {
	ID       string `json:"id"`
	UserName string `json:"userName,omitempty"`
}

type PostUserRequest struct {
	Schemas     []string `json:"schemas"`
	UserName    string   `json:"userName"`
	Id          string   `json:"id,omitempty"`
	Name        Name     `json:"name"`
	Emails      []Emails `json:"emails"`
	DisplayName string   `json:"displayName"`
	Locale      string   `json:"locale"`
	ExternalID  string   `json:"externalId"`
	Groups      []any    `json:"groups"`
	Password    string   `json:"password"`
	Active      bool     `json:"active"`
}
type Name struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}
type Emails struct {
	Primary bool   `json:"primary"`
	Value   string `json:"value"`
	Type    string `json:"type"`
}
