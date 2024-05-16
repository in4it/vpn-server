package license

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/in4it/wireguard-server/pkg/storage"
)

const AWS_PRODUCT_CODE = "7h7h3bnutjn0ziamv7npi8a69"

func getMetadataToken(client http.Client) string {
	metadataEndpoint := "http://" + metadataIP + "/latest/api/token"

	req, err := http.NewRequest("PUT", metadataEndpoint, nil)
	if err != nil {
		return ""
	}

	req.Header.Add("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return ""
}

func isOnAWSMarketPlace(client http.Client) bool {
	token := getMetadataToken(client)

	instanceIdentityDocument, err := getInstanceIdentityDocument(client, token)
	if err != nil {
		return false
	}
	for _, productCode := range instanceIdentityDocument.MarketplaceProductCodes {
		if productCode == AWS_PRODUCT_CODE {
			return true
		}
	}
	return false
}

func isOnAWS(client http.Client) bool {
	token := getMetadataToken(client)

	instanceIdentityDocument, err := getInstanceIdentityDocument(client, token)
	if err != nil {
		return false
	}
	return instanceIdentityDocument.AccountID != "" || instanceIdentityDocument.Version != ""
}

func getInstanceIdentityDocument(client http.Client, token string) (InstanceIdentityDocument, error) {
	var instanceIdentityDocument InstanceIdentityDocument

	endpoint := "http://" + metadataIP + "/2022-09-24/dynamic/instance-identity/document"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return instanceIdentityDocument, err
	}
	if token != "" {
		req.Header.Add("X-aws-ec2-metadata-token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return instanceIdentityDocument, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return instanceIdentityDocument, err
	}
	err = json.NewDecoder(resp.Body).Decode(&instanceIdentityDocument)
	if err != nil {
		return instanceIdentityDocument, err
	}

	return instanceIdentityDocument, nil
}

func GetMaxUsersAWSBYOL(client http.Client, storage storage.ReadWriter) int {
	userLicense := 3
	licenseKey, err := getAWSLicenseKey(storage, client)
	if err != nil {
		return userLicense
	}
	license, err := getLicense(client, licenseKey)
	if err != nil {
		return userLicense
	}
	return license.Users
}

func getAWSLicenseKey(storage storage.ReadWriter, client http.Client) (string, error) {
	token := getMetadataToken(client)
	licenseKey, err := getLicenseFromMetaData(token, client)
	if err != nil || licenseKey == "" {
		licenseKey, err = getLicenseKeyFromFile(storage)
		if err != nil {
			return "", err
		}
	}

	instanceIdentityDocument, err := getInstanceIdentityDocument(client, token)
	if err != nil {
		return "", err
	}

	return generateLicenseKey(licenseKey, instanceIdentityDocument.AccountID), nil
}

func getLicense(client http.Client, key string) (License, error) {
	var license License
	endpoint := licenseURL + "/" + key
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return license, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return license, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return license, fmt.Errorf("statuscode %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&license)
	if err != nil {
		return license, err
	}

	return license, nil

}

func getLicenseFromMetaData(token string, client http.Client) (string, error) {
	endpoint := "http://" + metadataIP + "/2022-09-24/meta-data/tags/instance/license"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Add("X-aws-ec2-metadata-token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", err
	}
	return string(bodyBytes), nil
}

func getAWSInstanceType(client http.Client) string {
	token := getMetadataToken(client)

	endpoint := "http://" + metadataIP + "/latest/meta-data/instance-type"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return ""
	}
	if token != "" {
		req.Header.Add("X-aws-ec2-metadata-token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return ""
}

func GetMaxUsersAWS(instanceType string) int {
	if instanceType == "" {
		return 3
	}
	if strings.HasSuffix(instanceType, ".nano") {
		return 3
	}
	if strings.HasSuffix(instanceType, ".micro") {
		return 10
	}
	if strings.HasSuffix(instanceType, ".small") {
		return 25
	}
	if strings.HasSuffix(instanceType, ".medium") {
		return 50
	}
	if strings.HasSuffix(instanceType, ".large") {
		return 100
	}
	if strings.HasSuffix(instanceType, ".xlarge") {
		return 250
	}
	if strings.HasSuffix(instanceType, ".2xlarge") {
		return 500
	}
	if strings.HasSuffix(instanceType, ".4xlarge") {
		return 1000
	}
	if strings.HasSuffix(instanceType, ".8xlarge") {
		return 2500
	}
	if strings.HasSuffix(instanceType, ".12xlarge") {
		return 5000
	}
	if strings.HasSuffix(instanceType, ".16xlarge") {
		return 10000
	}
	if strings.HasSuffix(instanceType, ".24xlarge") {
		return 10000
	}
	if strings.HasSuffix(instanceType, ".32xlarge") {
		return 10000
	}
	if strings.HasSuffix(instanceType, ".48xlarge") {
		return 10000
	}
	if strings.HasSuffix(instanceType, ".metal") {
		return 10000
	}

	return 3
}
