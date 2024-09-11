package license

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	randomutils "github.com/in4it/wireguard-server/pkg/utils/random"
)

var MetadataIP = "169.254.169.254"
var licenseURL = "https://in4it-vpn-server.s3.amazonaws.com/licenses"

func guessInfrastructure() string {
	// check whether we are on AWS, Azure, DigitalOcean or something undefined
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	if isOnAWSMarketPlace(client) {
		return "aws-marketplace"
	}

	if isOnAWS(client) {
		return "aws"
	}

	if isOnAzure(client) {
		return "azure"
	}

	if isOnDigitalOcean(client) {
		return "digitalocean"
	}

	if isOnGCP(client) {
		return "gcp"
	}

	return "" // no metadata server found
}

func GetInstanceType() (string, string) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	switch guessInfrastructure() {
	case "azure":
		return "azure", getAzureInstanceType(client)
	case "aws-marketplace":
		return "aws-marketplace", getAWSInstanceType(client)
	case "aws":
		return "aws", getAWSInstanceType(client)
	case "digitalocean":
		return "digitalocean", "droplet"
	case "gcp":
		return "gcp", "instance"
	default:
		return "", ""
	}
}
func GetMaxUsers(storage storage.ReadWriter) (int, string) {
	cloudType, instanceType := GetInstanceType()
	return getMaxUsers(storage, cloudType, instanceType), cloudType
}
func getMaxUsers(storage storage.ReadWriter, cloudType, instanceType string) int {
	switch cloudType {
	case "azure":
		return GetMaxUsersAzure(instanceType)
	case "aws-marketplace":
		return GetMaxUsersAWS(instanceType)
	case "aws":
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		return GetMaxUsersAWSBYOL(client, storage)
	case "digitalocean":
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		return GetMaxUsersDigitalOceanBYOL(client, storage)
	case "":
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		return GetMaxUsersBYOLNoCloud(client, storage)
	default:
		return 3
	}
}

func RefreshLicense(storage storage.ReadWriter, cloudType string, currentLicense int) int {
	if cloudType == "azure" || cloudType == "aws-marketplace" { // instance types / license is not going to change without a restart
		return currentLicense
	}
	cloudType, instanceType := GetInstanceType()
	return getMaxUsers(storage, cloudType, instanceType)
}

func GetLicenseKey(storage storage.ReadWriter, cloudType string) string {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	switch cloudType {
	case "aws":
		licenseKey, err := getAWSLicenseKey(storage, client)
		if err != nil {
			logging.DebugLog(fmt.Errorf("getAWSLicense error: %s", err))
			return ""
		}
		return licenseKey
	case "digitalocean":
		licenseKey, err := getDigitalOceanLicenseKey(storage, client)
		if err != nil {
			logging.DebugLog(fmt.Errorf("getDigitalOceanLicense error: %s", err))
			return ""
		}
		return licenseKey
	case "gcp":
		licenseKey, err := getGCPLicenseKey(storage, client)
		if err != nil {
			logging.DebugLog(fmt.Errorf("getGCPLicenseKey error: %s", err))
			return ""
		}
		return licenseKey
	default:
		licenseKey, err := getLicenseKeyFromFile(storage)
		if err != nil {
			logging.DebugLog(fmt.Errorf("getLicenseKeyFromFile error: %s", err))
			return ""
		}
		return licenseKey
	}

}

func generateLicenseKey(key string, identifier string) string {
	h := sha256.New()
	h.Write([]byte(identifier))
	bs := h.Sum(nil)

	return key + "-" + fmt.Sprintf("%x", bs)
}

func getLicenseKeyFromFile(storage storage.ReadWriter) (string, error) {
	filename := storage.ConfigPath("license.key")

	if storage.FileExists(filename) {
		licenseKeyBytes, err := storage.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("License read error: %s", err)
		}
		return string(licenseKeyBytes), nil
	}
	key, err := randomutils.GetRandomString(128)
	if err != nil {
		return "", fmt.Errorf("License generation error: %s", err)
	}
	err = storage.WriteFile(filename, []byte(key))
	if err != nil {
		return "", fmt.Errorf("License read error: %s", err)
	}
	return key, nil
}
