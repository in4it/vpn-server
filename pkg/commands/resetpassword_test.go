package commands

import (
	"testing"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
	"github.com/in4it/wireguard-server/pkg/users"
)

func TestResetPassword(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	adminCreated, err := ResetPassword(storage, "mytestpassword")
	if err != nil {
		t.Fatalf("reset password error: %s", err)
	}
	if !adminCreated {
		t.Fatalf("expected newly user to be created, received an userdatabase update instead")
	}
	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("userstore initialization error: %s", err)
	}
	user, err := userStore.GetUserByLogin("admin")
	if err != nil {
		t.Fatalf("get user by loginerror: %s", err)
	}
	if user.Login != "admin" {
		t.Fatalf("retrieved user is not admin")
	}
	if _, authOK := userStore.AuthUser("admin", "mytestpassword"); !authOK {
		t.Fatalf("couldn't authenticate admin")
	}
}
func TestResetPasswordExistingAdmin(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("userstore initialization error: %s", err)
	}
	_, err = userStore.AddUser(users.User{ID: "1-2-3-4", Login: "admin", Role: "admin"})
	if err != nil {
		t.Fatalf("could not add user: %s", err)
	}

	adminCreated, err := ResetPassword(storage, "mytestpassword")
	if err != nil {
		t.Fatalf("reset password error: %s", err)
	}
	if adminCreated {
		t.Fatalf("expected admin user to already exist")
	}
	userStore, err = users.NewUserStore(storage, -1) // user store is not in sync anymore with the file
	if err != nil {
		t.Fatalf("userstore initialization error: %s", err)
	}
	if _, authOK := userStore.AuthUser("admin", "mytestpassword"); !authOK {
		t.Fatalf("couldn't authenticate admin")
	}
}

func TestResetPasswordExistingAdminResetMFA(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("userstore initialization error: %s", err)
	}
	factors := []users.Factor{
		{
			Name:   "google",
			Type:   "otp",
			Secret: "123456",
		},
	}
	_, err = userStore.AddUser(users.User{ID: "1-2-3-4", Login: "admin", Role: "admin", Factors: factors})
	if err != nil {
		t.Fatalf("could not add user: %s", err)
	}

	adminCreated, err := ResetPassword(storage, "mytestpassword")
	if err != nil {
		t.Fatalf("reset password error: %s", err)
	}
	if adminCreated {
		t.Fatalf("expected admin user to already exist")
	}
	err = ResetAdminMFA(storage)
	if err != nil {
		t.Fatalf("reset admin mfa error: %s", err)
	}
	userStore, err = users.NewUserStore(storage, -1) // user store is not in sync anymore with the file
	if err != nil {
		t.Fatalf("userstore initialization error: %s", err)
	}
	user, err := userStore.GetUserByLogin("admin")
	if err != nil {
		t.Fatalf("get user by login error: %s", err)
	}
	if len(user.Factors) > 0 {
		t.Fatalf("found MFA for admin user")
	}
}
