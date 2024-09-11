package license

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

func isOnAzure(client http.Client) bool {
	req, err := http.NewRequest("GET", "http://"+MetadataIP+"/metadata/versions", nil)
	if err != nil {
		return false
	}

	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func GetMaxUsersAzure(instanceType string) int {
	if instanceType == "" {
		return 3
	}
	// patterns
	versionPattern := regexp.MustCompile(`^.*v[0-9]+#`)
	cpuPattern := regexp.MustCompile("[0-9]+")

	// extract amount of CPUs
	instanceTypeNoVersion := versionPattern.ReplaceAllString(instanceType, "")

	instanceTypeCPUs := cpuPattern.FindAllString(instanceTypeNoVersion, -1)

	if len(instanceTypeCPUs) > 0 {
		instanceTypeCPUCount, err := strconv.Atoi(instanceTypeCPUs[0])
		if err != nil {
			return 3
		}
		if instanceTypeCPUCount == 0 {
			return 15
		}
		return instanceTypeCPUCount * 25
	}

	return 3
}
func getAzureInstanceType(client http.Client) string {
	metadataEndpoint := "http://" + MetadataIP + "/metadata/instance?api-version=2021-02-01"
	req, err := http.NewRequest("GET", metadataEndpoint, nil)
	if err != nil {
		return ""
	}

	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var instanceMetadata AzureInstanceMetadata
	err = json.Unmarshal(bodyBytes, &instanceMetadata)
	if err != nil {
		return ""
	}
	return instanceMetadata.Compute.VMSize
}
