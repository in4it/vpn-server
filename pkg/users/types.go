package users

import "github.com/in4it/wireguard-server/pkg/storage"

type UserStore struct {
	Users    []User `json:"users"`
	autoSave bool
	maxUsers int
	storage  storage.Iface
}

type User struct {
	ID                               string   `json:"id"`
	Login                            string   `json:"login"`
	Role                             string   `json:"role"`
	OIDCID                           string   `json:"oidcID,omitempty"`
	SAMLID                           string   `json:"samlID,omitempty"`
	Provisioned                      bool     `json:"provisioned,omitempty"`
	Password                         string   `json:"password,omitempty"`
	Suspended                        bool     `json:"suspended"`
	ConnectionsDisabledOnAuthFailure bool     `json:"connectionsDisabledOnAuthFailure"`
	Factors                          []Factor `json:"factors"`
	ExternalID                       string   `json:"externalID,omitempty"`
}
type Factor struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Secret string `json:"secret"`
}
