package license

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestGetMaxUsersAzure(t *testing.T) {
	usersPerVCPU := 25
	testCases := map[string]int{
		"Standard_B1s":       usersPerVCPU,
		"Basic_A0":           15,
		"Standard_D1_v2":     usersPerVCPU,
		"Standard_D5_v2":     usersPerVCPU * 5,
		"D96as_v6":           usersPerVCPU * 96,
		"Standard_D16pls_v5": usersPerVCPU * 16,
		"Standard_DC1s_v3":   usersPerVCPU * 1,
	}
	for k, v := range testCases {
		if GetMaxUsersAzure(k) != v {
			t.Fatalf("Wrong output: %d vs %d", GetMaxUsersAzure(k), v)
		}
	}
}

func TestGetMaxUsersAWSMarketplace(t *testing.T) {
	testCases := map[string]int{
		"t3.medium": 50,
		"t3.large":  100,
		"t3.xlarge": 250,
	}
	for instanceType, v := range testCases {
		if getMaxUsers(&memorystorage.MockMemoryStorage{}, "aws-marketplace", instanceType) != v {
			t.Fatalf("Wrong output: %d vs %d", GetMaxUsersAWS(instanceType), v)
		}
	}
}

func TestGetMaxUsersAWSBYOL(t *testing.T) {
	accountID := "1234567890"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/versions" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.RequestURI == "/latest/api/token" {
			w.Write([]byte("this is a test token"))
			return
		}
		if r.RequestURI == "/2022-09-24/dynamic/instance-identity/document" {
			w.Write([]byte(`{
  "accountId" : "` + accountID + `",
  "architecture" : "x86_64",
  "availabilityZone" : "us-east-1c",
  "billingProducts" : null,
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : [ "7h7h3bnutjn0ziamv7npi8a69" ],
  "imageId" : "ami-12345678",
  "instanceId" : "i-123456",
  "instanceType" : "t3.micro",
  "kernelId" : null,
  "pendingTime" : "2024-06-15T08:34:50Z",
  "privateIp" : "10.0.1.123",
  "ramdiskId" : null,
  "region" : "us-east-1",
  "version" : "2017-09-30"
}`))

			return
		}
		if r.RequestURI == "/2022-09-24/meta-data/tags/instance/license" {
			w.Write([]byte(`license-1234556-license`))
			return
		}
		h := sha256.New()
		h.Write([]byte(accountID))
		if r.RequestURI == fmt.Sprintf("/license-1234556-license-%x", h.Sum(nil)) {
			w.Write([]byte(`{"users": 50}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	testCases := map[string]int{
		"t3.medium": 50,
		"t3.xlarge": 50,
	}
	licenseURL = ts.URL
	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)
	for _, v := range testCases {
		if v2 := GetMaxUsersAWSBYOL(http.Client{Timeout: 5 * time.Second}, &memorystorage.MockMemoryStorage{}); v2 != v {
			t.Fatalf("Wrong output: %d vs %d", v2, v)
		}
	}
}

func TestGetMaxUsersAWS(t *testing.T) {
	testCases := map[string]int{
		"t3.medium": 3,
		"t3.large":  3,
		"t3.xlarge": 3,
	}
	for instanceType, v := range testCases {
		if getMaxUsers(&memorystorage.MockMemoryStorage{}, "aws", instanceType) != v {
			t.Fatalf("Wrong output: %d vs %d", GetMaxUsersAWS(instanceType), v)
		}
	}
}

func TestGuessInfrastructureAzure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/versions" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "azure" {
		t.Fatalf("wrong infra returned: %s", infra)
	}
}

func TestGuessInfrastructureAWSMarketplace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/versions" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.RequestURI == "/latest/api/token" {
			w.Write([]byte("this is a test token"))
			return
		}
		if r.RequestURI == "/2022-09-24/dynamic/instance-identity/document" {
			w.Write([]byte(`{
  "accountId" : "12345678",
  "architecture" : "x86_64",
  "availabilityZone" : "us-east-1c",
  "billingProducts" : null,
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : [ "7h7h3bnutjn0ziamv7npi8a69" ],
  "imageId" : "ami-12345678",
  "instanceId" : "i-123456",
  "instanceType" : "t3.micro",
  "kernelId" : null,
  "pendingTime" : "2024-06-15T08:34:50Z",
  "privateIp" : "10.0.1.123",
  "ramdiskId" : null,
  "region" : "us-east-1",
  "version" : "2017-09-30"
}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "aws-marketplace" {
		t.Fatalf("wrong infra returned: %s", infra)
	}
}

func TestGuessInfrastructureAWS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/versions" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.RequestURI == "/latest/api/token" {
			w.Write([]byte("this is a test token"))
			return
		}
		if r.RequestURI == "/2022-09-24/dynamic/instance-identity/document" {
			w.Write([]byte(`{
  "accountId" : "12345678",
  "architecture" : "x86_64",
  "availabilityZone" : "us-east-1c",
  "billingProducts" : null,
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : null
}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "aws" {
		t.Fatalf("wrong infra returned: %s", infra)
	}

	if getMaxUsers(&memorystorage.MockMemoryStorage{}, infra, "t3.large") != 3 {
		t.Fatalf("wrong users returned")
	}
}

func TestGuessInfrastructureOther(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "" {
		t.Fatalf("wrong infra returned: %s", infra)
	}
}

func TestGetAzureInstanceType(t *testing.T) {
	vmSize := "Standard_D2_v5"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(AzureInstanceMetadata{
			Compute: Compute{
				VMSize: vmSize,
			},
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(out)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	usersPerVCPU := 25

	users := getMaxUsers(&memorystorage.MockMemoryStorage{}, "azure", getAzureInstanceType(http.Client{Timeout: 5 * time.Second}))

	if users != usersPerVCPU*2 {
		t.Fatalf("Wrong user count returned")
	}
}

func TestGetAWSInstanceType(t *testing.T) {
	instanceType := "t4.xlarge"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/latest/api/token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.RequestURI == "/latest/meta-data/instance-type" {
			w.Write([]byte(instanceType))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)

	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	users := GetMaxUsersAWS(getAWSInstanceType(http.Client{Timeout: 5 * time.Second}))

	if users != 250 {
		t.Fatalf("Wrong user count returned: %d", users)
	}
}

func TestGuessInfrastructureDigitalOcean(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/v1/interfaces/private/0/type" {
			w.Write([]byte(`private`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	infra := guessInfrastructure()

	if infra != "digitalocean" {
		t.Fatalf("wrong infra returned: %s", infra)
	}

	if getMaxUsers(&memorystorage.MockMemoryStorage{}, infra, "t3.large") != 3 {
		t.Fatalf("wrong users returned")
	}
}

func TestGetMaxUsersDigitalOceanBYOL(t *testing.T) {
	dropletID := "1234567890"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/v1/interfaces/private/0/type" {
			w.Write([]byte(`private`))
			return
		}
		if r.RequestURI == "/metadata/v1/id" {
			w.Write([]byte(dropletID))

			return
		}
		h := sha256.New()
		h.Write([]byte(dropletID))
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
		if v2 := GetMaxUsersDigitalOceanBYOL(http.Client{Timeout: 5 * time.Second}, mockStorage); v2 != v {
			t.Fatalf("Wrong output: %d vs %d", v2, v)
		}
	}
}

func TestGetLicenseKey(t *testing.T) {
	dropletID := "1234567890"
	projectID := "googleproject-12356"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/v1/interfaces/private/0/type" {
			w.Write([]byte(`private`))
			return
		}
		if r.RequestURI == "/metadata/v1/id" {
			w.Write([]byte(dropletID))
			return
		}
		if r.RequestURI == "/computeMetadata/v1/project/project-id" {
			w.Write([]byte(projectID))
			return
		}
		if r.RequestURI == "/metadata/versions" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.RequestURI == "/latest/api/token" {
			w.Write([]byte("this is a test token"))
			return
		}
		if r.RequestURI == "/2022-09-24/dynamic/instance-identity/document" {
			w.Write([]byte(`{
  "accountId" : "12345678",
  "architecture" : "x86_64",
  "availabilityZone" : "us-east-1c",
  "billingProducts" : null,
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : null
}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	MetadataIP = strings.Replace(ts.URL, "http://", "", -1)

	logging.Loglevel = logging.LOG_DEBUG + logging.LOG_ERROR
	key := GetLicenseKey(&memorystorage.MockMemoryStorage{}, "")
	if key == "" {
		t.Fatalf("key is empty")
	}
	key = GetLicenseKey(&memorystorage.MockMemoryStorage{}, "aws")
	if key == "" {
		t.Fatalf("aws key is empty")
	}
	key = GetLicenseKey(&memorystorage.MockMemoryStorage{}, "digitalocean")
	if key == "" {
		t.Fatalf("digitalocean key is empty")
	}
	key = GetLicenseKey(&memorystorage.MockMemoryStorage{}, "gcp")
	if key == "" {
		t.Fatalf("gcp key is empty")
	}
}
func TestGetLicenseKeyNoCloudProvider(t *testing.T) {

	logging.Loglevel = logging.LOG_DEBUG + logging.LOG_ERROR
	key := GetLicenseKey(&memorystorage.MockMemoryStorage{}, "")
	if key == "" {
		t.Fatalf("key is empty")
	}
}
