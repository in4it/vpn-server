package login

import "github.com/in4it/wireguard-server/pkg/users"

type AuthIface interface {
	AuthUser(login string, password string) (users.User, bool)
}

type LoginRequest struct {
	Login          string         `json:"login"`
	Password       string         `json:"password"`
	FactorResponse FactorResponse `json:"factorResponse"`
}

type FactorResponse struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type LoginResponse struct {
	Authenticated bool     `json:"authenticated"`
	Suspended     bool     `json:"suspended"`
	NoLicense     bool     `json:"noLicense"`
	Token         string   `json:"token,omitempty"`
	MFARequired   bool     `json:"mfaRequired"`
	Factors       []string `json:"factors"`
}
