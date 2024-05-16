package oidcrenewal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
)

func refreshToken(discovery oidc.Discovery, refreshToken, clientID, clientSecret string) (oidc.Token, time.Time, error) {
	var token oidc.Token
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	if discovery.TokenEndpoint == "" {
		return token, time.Time{}, fmt.Errorf("token endpoint is empty")
	}

	payload := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	resp, err := client.PostForm(discovery.TokenEndpoint, payload)
	if err != nil {
		return token, time.Time{}, fmt.Errorf("tokenEndpoint PostForm error: %s", err)
	}
	renewalTime := time.Now()
	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return token, renewalTime, fmt.Errorf("tokenEndpoint return error. statuscode: %d", resp.StatusCode)
		}
		return token, renewalTime, fmt.Errorf("tokenEndpoint return error (statuscode %d): %s", resp.StatusCode, data)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		return token, renewalTime, fmt.Errorf("tokenEndpoint decode error: %s", err)
	}
	if token.AccessToken == "" {
		return token, renewalTime, fmt.Errorf("access token is empty")
	}

	return token, renewalTime, nil

}
