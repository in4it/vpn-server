package login

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"testing"
	"time"

	"github.com/in4it/wireguard-server/pkg/mfa/totp"
	"github.com/in4it/wireguard-server/pkg/users"
)

type MockAuth struct {
	AuthUserUser   users.User
	AuthUserResult bool
}

func (m *MockAuth) AuthUser(login string, password string) (users.User, bool) {
	return m.AuthUserUser, m.AuthUserResult
}

func TestAuthenticate(t *testing.T) {
	m := MockAuth{
		AuthUserUser: users.User{
			Login: "john",
		},
		AuthUserResult: true,
	}
	loginReq := LoginRequest{
		Login:    "john",
		Password: "mypass",
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("private key error: %s", err)
	}

	loginResp, _, err := Authenticate(loginReq, &m, privateKey, "jwtKeyID")
	if err != nil {
		t.Fatalf("authentication error: %s", err)
	}
	if !loginResp.Authenticated {
		t.Fatalf("expected to be authenticated")
	}
	if loginResp.Token == "" {
		t.Fatalf("no token")
	}
}
func TestAuthenticateMFANoToken(t *testing.T) {
	m := MockAuth{
		AuthUserUser: users.User{
			Login: "john",
			Factors: []users.Factor{
				{
					Name:   "test-factor",
					Type:   "test",
					Secret: "secret",
				},
			},
		},
		AuthUserResult: true,
	}
	loginReq := LoginRequest{
		Login:    "john",
		Password: "mypass",
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("private key error: %s", err)
	}

	loginResp, _, err := Authenticate(loginReq, &m, privateKey, "jwtKeyID")
	if err != nil {
		t.Fatalf("authentication error: %s", err)
	}
	if loginResp.Authenticated {
		t.Fatalf("expected not to be authenticated")
	}
	if len(loginResp.Factors) == 0 {
		t.Fatalf("expected to get factors")
	}
}
func TestAuthenticateMFAWithToken(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("secret"))
	m := MockAuth{
		AuthUserUser: users.User{
			Login: "john",
			Factors: []users.Factor{
				{
					Name:   "test-factor",
					Type:   "test",
					Secret: secret,
				},
			},
		},
		AuthUserResult: true,
	}
	token, err := totp.GetToken(secret, time.Now().Unix()/30)
	if err != nil {
		t.Fatalf("GetToken error: %s", err)
	}
	loginReq := LoginRequest{
		Login:    "john",
		Password: "mypass",
		FactorResponse: FactorResponse{
			Name: "test-factor",
			Code: token,
		},
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("private key error: %s", err)
	}

	loginResp, _, err := Authenticate(loginReq, &m, privateKey, "jwtKeyID")
	if err != nil {
		t.Fatalf("authentication error: %s", err)
	}
	if !loginResp.Authenticated {
		t.Fatalf("expected not to be authenticated")
	}
}
