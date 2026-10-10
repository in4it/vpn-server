package wireguard

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// cloud provider metadata endpoints, variables so tests can override them
var (
	awsMetadataURL          = "http://169.254.169.254"
	gcpMetadataURL          = "http://metadata.google.internal"
	azureMetadataURL        = "http://169.254.169.254"
	digitalOceanMetadataURL = "http://169.254.169.254"
	metadataTimeout         = 3 * time.Second
)

type publicIPLookup func(ctx context.Context, client *http.Client) (string, error)

// getPublicIPFromMetadata queries the metadata services of the supported cloud
// providers concurrently and returns the public ip of the first provider (in
// order of priority: AWS, GCP, Azure, DigitalOcean) that returns one. Returns
// an empty string when no public ip could be found.
func getPublicIPFromMetadata() string {
	ctx, cancel := context.WithTimeout(context.Background(), metadataTimeout)
	defer cancel()
	client := &http.Client{Timeout: metadataTimeout}

	lookups := []publicIPLookup{getAWSPublicIP, getGCPPublicIP, getAzurePublicIP, getDigitalOceanPublicIP}
	results := make([]chan string, len(lookups))
	for k, lookup := range lookups {
		results[k] = make(chan string, 1)
		go func(lookup publicIPLookup, result chan<- string) {
			ip, err := lookup(ctx, client)
			if err != nil {
				result <- ""
				return
			}
			result <- ip
		}(lookup, results[k])
	}
	for _, result := range results {
		if ip := <-result; ip != "" {
			return ip
		}
	}
	return ""
}

func getAWSPublicIP(ctx context.Context, client *http.Client) (string, error) {
	// IMDSv2 token (optional: IMDSv1 is tried when no token is returned)
	token := ""
	tokenBody, statusCode, err := metadataRequest(ctx, client, http.MethodPut, awsMetadataURL+"/latest/api/token", map[string]string{"X-aws-ec2-metadata-token-ttl-seconds": "21600"})
	if err != nil {
		return "", err
	}
	if statusCode == http.StatusOK {
		token = strings.TrimSpace(string(tokenBody))
	}
	headers := map[string]string{}
	if token != "" {
		headers["X-aws-ec2-metadata-token"] = token
	}
	return metadataIPRequest(ctx, client, awsMetadataURL+"/latest/meta-data/public-ipv4", headers)
}

func getGCPPublicIP(ctx context.Context, client *http.Client) (string, error) {
	return metadataIPRequest(ctx, client, gcpMetadataURL+"/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip", map[string]string{"Metadata-Flavor": "Google"})
}

func getAzurePublicIP(ctx context.Context, client *http.Client) (string, error) {
	return metadataIPRequest(ctx, client, azureMetadataURL+"/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2021-02-01&format=text", map[string]string{"Metadata": "true"})
}

func getDigitalOceanPublicIP(ctx context.Context, client *http.Client) (string, error) {
	return metadataIPRequest(ctx, client, digitalOceanMetadataURL+"/metadata/v1/interfaces/public/0/ipv4/address", nil)
}

// metadataIPRequest does a GET request and returns the body if it's a valid ip address
func metadataIPRequest(ctx context.Context, client *http.Client, url string, headers map[string]string) (string, error) {
	body, statusCode, err := metadataRequest(ctx, client, http.MethodGet, url, headers)
	if err != nil {
		return "", err
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("metadata request returned status code %d", statusCode)
	}
	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("metadata request didn't return a valid ip address")
	}
	return ip, nil
}

func metadataRequest(ctx context.Context, client *http.Client, method, url string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}
