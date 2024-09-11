package license

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestGuessInfrastructureGCP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/computeMetadata/v1/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "gcp" {
		t.Fatalf("wrong infra returned: %s", infra)
	}
}

func TestGetMaxUsersGCPBYOL(t *testing.T) {
	projectID := "gcpproject-1234567890"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/computeMetadata/v1/project/project-id" {
			w.Write([]byte(projectID))

			return
		}
		h := sha256.New()
		h.Write([]byte(projectID))
		if r.RequestURI == fmt.Sprintf("/license-1234556-license-%x", h.Sum(nil)) {
			w.Write([]byte(`{"users": 50}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	licenseURL = ts.URL
	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	mockStorage := &memorystorage.MockMemoryStorage{}
	err := mockStorage.WriteFile("config/license.key", []byte("license-1234556-license"))
	if err != nil {
		t.Fatalf("writefile error: %s", err)
	}

	for _, v := range []int{50} {
		if v2 := GetMaxUsersGCPBYOL(http.Client{Timeout: 5 * time.Second}, mockStorage); v2 != v {
			t.Fatalf("Wrong output: %d vs %d", v2, v)
		}
	}
}
