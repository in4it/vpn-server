package rest

import (
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	oidcstore "github.com/in4it/wireguard-server/pkg/auth/oidc/store"
	oidcrenewal "github.com/in4it/wireguard-server/pkg/auth/oidc/store/renewal"
	"github.com/in4it/wireguard-server/pkg/auth/provisioning/scim"
	"github.com/in4it/wireguard-server/pkg/auth/saml"
	"github.com/in4it/wireguard-server/pkg/observability"
	"github.com/in4it/wireguard-server/pkg/rest/login"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
)

const SETUP_CODE_FILE = "setup-code.txt"
const ADMIN_USER = "admin"

type Context struct {
	AppDir                  string               `json:"appDir,omitempty"`
	ServerType              string               `json:"serverType,omitempty"`
	SetupCompleted          bool                 `json:"setupCompleted"`
	Hostname                string               `json:"hostname,omitempty"`
	Protocol                string               `json:"protocol,omitempty"`
	JWTKeys                 *JWTKeys             `json:"jwtKeys,omitempty"`
	JWTKeysKID              string               `json:"jwtKeysKid,omitempty"`
	OIDCProviders           []oidc.OIDCProvider  `json:"oidcProviders,omitempty"`
	LocalAuthDisabled       bool                 `json:"disableLocalAuth,omitempty"`
	EnableTLS               bool                 `json:"enableTLS,omitempty"`
	RedirectToHttps         bool                 `json:"redirectToHttps,omitempty"`
	EnableOIDCTokenRenewal  bool                 `json:"enableOIDCTokenRenewal,omitempty"`
	OIDCStore               *oidcstore.Store     `json:"oidcStore,omitempty"`
	UserStore               *users.UserStore     `json:"users,omitempty"`
	OIDCRenewal             *oidcrenewal.Renewal `json:"oidcRenewal,omitempty"`
	LoginAttempts           login.Attempts       `json:"loginAttempts,omitempty"`
	LicenseUserCount        int                  `json:"licenseUserCount,omitempty"`
	CloudType               string               `json:"cloudType,omitempty"`
	TokenRenewalTimeMinutes int                  `json:"tokenRenewalTimeMinutes,omitempty"`
	LogLevel                int                  `json:"loglevel,omitempty"`
	SCIM                    *SCIM                `json:"scim,omitempty"`
	SAML                    *SAML                `json:"saml,omitempty"`
	Observability           *Observability       `json:"observability,omitempty"`
	Storage                 *Storage             `json:"storage,omitempty"`
}
type SCIM struct {
	EnableSCIM bool       `json:"enableSCIM,omitempty"`
	Token      string     `json:"token"`
	Client     scim.Iface `json:"client,omitempty"`
}
type SAML struct {
	Providers *[]saml.Provider `json:"providers"`
	Client    saml.Iface       `json:"client,omitempty"`
}
type Observability struct {
	Client observability.Iface `json:"client,omitempty"`
}
type Storage struct {
	Client storage.Iface `json:"client,omitempty"`
}

type ContextRequest struct {
	Secret        string `json:"secret"`
	AdminPassword string `json:"adminPassword"`
	Hostname      string `json:"hostname"`
	Protocol      string `json:"protocol"`
}

type AuthMethodsResponse struct {
	LocalAuthDisabled bool                  `json:"localAuthDisabled"`
	OIDCProviders     []AuthMethodsProvider `json:"oidcProviders"`
}

type AuthMethodsProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RedirectURI string `json:"redirectURI,omitempty"`
}

type OIDCCallback struct {
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectURI string `json:"redirectURI"`
}
type SAMLCallback struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirectURI"`
}

type UserInfoResponse struct {
	Login    string `json:"login"`
	Role     string `json:"role"`
	UserType string `json:"userType"`
}

type GeneralSetupRequest struct {
	Hostname               string `json:"hostname"`
	EnableTLS              bool   `json:"enableTLS"`
	RedirectToHttps        bool   `json:"redirectToHttps"`
	DisableLocalAuth       bool   `json:"disableLocalAuth"`
	EnableOIDCTokenRenewal bool   `json:"enableOIDCTokenRenewal"`
}

type VPNSetupRequest struct {
	Routes              string `json:"routes"`
	VPNEndpoint         string `json:"vpnEndpoint"`
	AddressRange        string `json:"addressRange"`
	ClientAddressPrefix string `json:"clientAddressPrefix"`
	Port                string `json:"port"`
	ExternalInterface   string `json:"externalInterface"`
	Nameservers         string `json:"nameservers"`
	DisableNAT          bool   `json:"disableNAT"`
}

type TemplateSetupRequest struct {
	ClientTemplate string `json:"clientTemplate"`
	ServerTemplate string `json:"serverTemplate"`
}

type NewConnectionResponse struct {
	Name string `json:"name"`
}
type Connection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LicenseResponse struct {
	LicenseUserCount int    `json:"licenseUserCount"`
	CurrentUserCount int    `json:"currentUserCount,omitempty"`
	CloudType        string `json:"cloudType"`
	Key              string `json:"key,omitempty"`
}

type ConnectionLicenseResponse struct {
	LicenseUserCount int `json:"licenseUserCount"`
	ConnectionCount  int `json:"connectionCount"`
}

type JwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

type UsersResponse struct {
	ID                               string    `json:"id"`
	Login                            string    `json:"login"`
	Role                             string    `json:"role"`
	OIDCID                           string    `json:"oidcID"`
	SAMLID                           string    `json:"samlID"`
	Provisioned                      bool      `json:"provisioned"`
	Suspended                        bool      `json:"suspended"`
	ConnectionsDisabledOnAuthFailure bool      `json:"connectionsDisabledOnAuthFailure"`
	LastTokenRenewal                 time.Time `json:"lastTokenRenewal,omitempty"`
	LastLogin                        string    `json:"lastLogin"`
}

type FactorRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

type SCIMSetup struct {
	Enabled         bool   `json:"enabled"`
	Token           string `json:"token,omitempty"`
	RegenerateToken bool   `json:"regenerateToken,omitempty"`
	BaseURL         string `json:"baseURL,omitempty"`
}

type SAMLSetup struct {
	Enabled        bool   `json:"enabled"`
	MetadataURL    string `json:"metadataURL,omitempty"`
	RegenerateCert bool   `json:"regenerateCert,omitempty"`
}

type UserStatsResponse struct {
	ReceiveBytes  UserStatsData `json:"receivedBytes"`
	TransmitBytes UserStatsData `json:"transmitBytes"`
	Handshakes    UserStatsData `json:"handshakes"`
}
type UserStatsData struct {
	Datasets UserStatsDatasets `json:"datasets"`
}
type UserStatsDatasets []UserStatsDataset
type UserStatsDataset struct {
	Label           string               `json:"label"`
	Data            []UserStatsDataPoint `json:"data"`
	Fill            bool                 `json:"fill"`
	BorderColor     string               `json:"borderColor"`
	BackgroundColor string               `json:"backgroundColor"`
	Tension         float64              `json:"tension"`
	ShowLine        bool                 `json:"showLine"`
}

type UserStatsDataPoint struct {
	X string  `json:"x"`
	Y float64 `json:"y"`
}

type NewUserRequest struct {
	Login    string `json:"login"`
	Role     string `json:"role"`
	Password string `json:"password,omitempty"`
}
