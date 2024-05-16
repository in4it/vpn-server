package oidc

import (
	"fmt"
	"strings"
)

func GetRedirectURI(discovery Discovery, clientID, scope, callback string, enableOIDCTokenRenewal bool) (string, string, error) {
	var redirectURI string

	state, err := GetRandomString(64)
	if err != nil {
		return redirectURI, state, fmt.Errorf("GetRandomString error: %s", err)
	}

	// add offline_access to scope if oidc token renewal is true
	//if enableOIDCTokenRenewal {
	//	scope = strings.TrimSpace(scope) + " offline_access"
	//}

	scope = strings.Replace(scope, " ", "%20", -1)

	redirectURI = fmt.Sprintf("%s?client_id=%s&state=%s&scope=%s&response_type=code&redirect_uri=%s", discovery.AuthorizationEndpoint, clientID, state, scope, callback)

	return redirectURI, state, nil
}
