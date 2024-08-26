package login

import (
	"crypto/rsa"
	"fmt"

	"github.com/in4it/wireguard-server/pkg/mfa/totp"
	"github.com/in4it/wireguard-server/pkg/users"
)

func Authenticate(loginReq LoginRequest, authIface AuthIface, jwtPrivateKey *rsa.PrivateKey, jwtKeyID string) (LoginResponse, users.User, error) {
	loginResponse := LoginResponse{}
	user, auth := authIface.AuthUser(loginReq.Login, loginReq.Password)
	if auth && !user.Suspended {
		if len(user.Factors) == 0 { // authentication without MFA
			token, err := GetJWTToken(user.Login, user.Role, jwtPrivateKey, jwtKeyID)
			if err != nil {
				return loginResponse, user, fmt.Errorf("token generation failed: %s", err)
			}
			loginResponse.Authenticated = true
			loginResponse.Token = token
		} else {
			if loginReq.FactorResponse.Name == "" {
				loginResponse.Authenticated = false
				loginResponse.MFARequired = true
				for _, factor := range user.Factors {
					loginResponse.Factors = append(loginResponse.Factors, factor.Name)
				}
			} else {
				for _, factor := range user.Factors {
					if factor.Name == loginReq.FactorResponse.Name {
						ok, err := totp.Verify(factor.Secret, loginReq.FactorResponse.Code)
						if err != nil {
							return loginResponse, user, fmt.Errorf("MFA (totp) verify failed: %s", err)
						}
						if ok { // authentication with MFA
							token, err := GetJWTToken(user.Login, user.Role, jwtPrivateKey, jwtKeyID)
							if err != nil {
								return loginResponse, user, fmt.Errorf("token generation failed: %s", err)
							}
							loginResponse.Authenticated = true
							loginResponse.Token = token
						}
					}
				}
			}
		}
	}
	return loginResponse, user, nil
}
