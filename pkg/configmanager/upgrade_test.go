package configmanager

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestDownloadFilesForUpgrade(t *testing.T) {
	t.Skip() // test downloads are ad-hoc
	pwd, err := os.Executable()
	if err != nil {
		t.Fatalf("upgrade: user current dir error: %s", err)
	}
	pwdDir := path.Dir(pwd)
	err = downloadFilesForUpgrade(pwdDir, map[string]string{
		"rest-server":          "restserver-linux-amd64",
		"reset-admin-password": "reset-admin-password-linux-amd64",
		"configmanager":        "configmanager-linux-amd64",
	})
	if err != nil {
		t.Fatalf("upgrade error: %s", err)
	}
}

func TestNewVersionAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() == "/latest" {
			w.Write([]byte("v1.0.38"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	BINARIES_URL = server.URL

	available, version, err := newVersionAvailable()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if available {
		t.Fatalf("expected new version not to be available: %s", version)
	}
}

func TestNewVersionAvailableSameVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() == "/latest" {
			w.Write([]byte(getVersion()))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	BINARIES_URL = server.URL

	available, version, err := newVersionAvailable()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if available {
		t.Fatalf("expected new version not to be available: %s", version)
	}
}

func TestNewVersionAvailableHigherVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() == "/latest" {
			currentVersionSplit := strings.Split(getVersion(), ".")
			if len(currentVersionSplit) != 3 {
				t.Fatalf("unsupported current version: %s", getVersion())
			}
			i, err := strconv.Atoi(currentVersionSplit[2])
			if err != nil {
				t.Fatalf("unsupported current version: %s", getVersion())
			}
			i++
			newVersion := strings.Join([]string{currentVersionSplit[0], currentVersionSplit[1], strconv.Itoa(i)}, ".")
			w.Write([]byte(newVersion))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	BINARIES_URL = server.URL

	available, version, err := newVersionAvailable()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if !available {
		t.Fatalf("expected new version expected to be available: %s", version)
	}
}
func TestNewVersionAvailableBogus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() == "/latest" {
			w.Write([]byte("v1.x.38"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	BINARIES_URL = server.URL

	available, version, err := newVersionAvailable()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if available {
		t.Fatalf("expected new version not to be available: %s", version)
	}
}

func TestNewVersionAvailableBogus2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() == "/latest" {
			w.Write([]byte("v2"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	BINARIES_URL = server.URL

	available, version, err := newVersionAvailable()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if available {
		t.Fatalf("expected new version not to be available: %s", version)
	}
}
