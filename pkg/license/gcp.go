package license

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
)

func isOnGCP(client http.Client) bool {
	endpoint := "http://" + metadataIP + "/computeMetadata/v1/"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false
	}

	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func GetMaxUsersGCPBYOL(client http.Client, storage storage.ReadWriter) int {
	userLicense := 3

	licenseKey, err := getGCPLicenseKey(storage, client)
	if err != nil {
		logging.DebugLog(fmt.Errorf("get gcp license error: %s", err))
		return userLicense
	}

	license, err := getLicense(client, licenseKey)
	if err != nil {
		logging.DebugLog(fmt.Errorf("getLicense error: %s", err))
		return userLicense
	}

	return license.Users
}

func getGCPLicenseKey(storage storage.ReadWriter, client http.Client) (string, error) {
	identifier, err := getGCPIdentifier(client)
	if err != nil {
		logging.DebugLog(fmt.Errorf("License generation error (identifier error): %s", err))
		return "", err
	}

	licenseKey, err := getLicenseKeyFromFile(storage)
	if err != nil {
		return "", err
	}

	return generateLicenseKey(licenseKey, identifier), nil
}

func getGCPIdentifier(client http.Client) (string, error) {
	id := ""
	endpoint := "http://" + metadataIP + "/computeMetadata/v1/project/project-id"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return id, err
	}

	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return id, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return id, err
	}
	if resp.StatusCode != 200 {
		return id, fmt.Errorf("wrong statuscode returned: %d; body: %s", resp.StatusCode, body)
	}

	return strings.TrimSpace(string(body)), nil

}
