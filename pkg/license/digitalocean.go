package license

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
)

func isOnDigitalOcean(client http.Client) bool {
	endpoint := "http://" + MetadataIP + "/metadata/v1/interfaces/private/0/type"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func GetMaxUsersDigitalOceanBYOL(client http.Client, storage storage.ReadWriter) int {
	userLicense := 3

	licenseKey, err := getDigitalOceanLicenseKey(storage, client)
	if err != nil {
		logging.DebugLog(fmt.Errorf("get digitalocean license error: %s", err))
		return userLicense
	}

	license, err := getLicense(client, licenseKey)
	if err != nil {
		logging.DebugLog(fmt.Errorf("getLicense error: %s", err))
		return userLicense
	}

	return license.Users
}

func getDigitalOceanLicenseKey(storage storage.ReadWriter, client http.Client) (string, error) {
	identifier, err := getDigitalOceanIdentifier(client)
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

func getDigitalOceanIdentifier(client http.Client) (string, error) {
	id := ""
	endpoint := "http://" + MetadataIP + "/metadata/v1/id"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return id, err
	}

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

func HasDigitalOceanTagSet(client http.Client, tag string) (bool, error) {
	endpoint := "http://" + MetadataIP + "/metadata/v1/tags"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("wrong statuscode returned: %d; body: %s", resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if tag == strings.TrimSpace(scanner.Text()) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil

}
