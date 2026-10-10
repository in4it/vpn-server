package wireguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// metadataServer returns a test server that answers path (and requires the
// header, if set) with body
func metadataServer(t *testing.T, path, headerKey, headerValue, body string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if headerKey != "" && r.Header.Get(headerKey) != headerValue {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	return ts
}

func setMetadataURLs(t *testing.T, aws, gcp, azure, digitalOcean string) {
	t.Helper()
	origAWS, origGCP, origAzure, origDO, origTimeout := awsMetadataURL, gcpMetadataURL, azureMetadataURL, digitalOceanMetadataURL, metadataTimeout
	t.Cleanup(func() {
		awsMetadataURL, gcpMetadataURL, azureMetadataURL, digitalOceanMetadataURL, metadataTimeout = origAWS, origGCP, origAzure, origDO, origTimeout
	})
	awsMetadataURL, gcpMetadataURL, azureMetadataURL, digitalOceanMetadataURL = aws, gcp, azure, digitalOcean
	metadataTimeout = 2 * time.Second
}

func notFoundServer(t *testing.T) string {
	return metadataServer(t, "/none", "", "", "").URL
}

const gcpExternalIPPath = "/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"
const azurePublicIPPath = "/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2021-02-01&format=text"

func TestGetPublicIPFromMetadataGCP(t *testing.T) {
	gcp := metadataServer(t, gcpExternalIPPath, "Metadata-Flavor", "Google", "34.1.2.3\n")
	setMetadataURLs(t, notFoundServer(t), gcp.URL, notFoundServer(t), notFoundServer(t))
	if ip := getPublicIPFromMetadata(); ip != "34.1.2.3" {
		t.Fatalf("expected GCP external ip, got %q", ip)
	}
}

func TestGetPublicIPFromMetadataAzure(t *testing.T) {
	azure := metadataServer(t, azurePublicIPPath, "Metadata", "true", "20.1.2.3")
	setMetadataURLs(t, notFoundServer(t), notFoundServer(t), azure.URL, notFoundServer(t))
	if ip := getPublicIPFromMetadata(); ip != "20.1.2.3" {
		t.Fatalf("expected Azure public ip, got %q", ip)
	}
}

func TestGetPublicIPFromMetadataDigitalOcean(t *testing.T) {
	do := metadataServer(t, "/metadata/v1/interfaces/public/0/ipv4/address", "", "", "167.1.2.3")
	setMetadataURLs(t, notFoundServer(t), notFoundServer(t), notFoundServer(t), do.URL)
	if ip := getPublicIPFromMetadata(); ip != "167.1.2.3" {
		t.Fatalf("expected DigitalOcean public ip, got %q", ip)
	}
}

func TestGetPublicIPFromMetadataAWSFirst(t *testing.T) {
	aws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/latest/api/token":
			_, _ = w.Write([]byte("token"))
		case r.Method == http.MethodGet && r.URL.Path == "/latest/meta-data/public-ipv4" && r.Header.Get("X-aws-ec2-metadata-token") == "token":
			_, _ = w.Write([]byte("52.1.2.3"))
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	t.Cleanup(aws.Close)
	gcp := metadataServer(t, gcpExternalIPPath, "Metadata-Flavor", "Google", "34.1.2.3")
	setMetadataURLs(t, aws.URL, gcp.URL, notFoundServer(t), notFoundServer(t))
	if ip := getPublicIPFromMetadata(); ip != "52.1.2.3" {
		t.Fatalf("expected AWS public ip, got %q", ip)
	}
}

func TestGetPublicIPFromMetadataNone(t *testing.T) {
	// GCP without the Metadata-Flavor header is refused, and non-ip responses are ignored
	gcp := metadataServer(t, gcpExternalIPPath, "Metadata-Flavor", "Wrong", "34.1.2.3")
	azure := metadataServer(t, azurePublicIPPath, "", "", "<html>not an ip</html>")
	setMetadataURLs(t, notFoundServer(t), gcp.URL, azure.URL, notFoundServer(t))
	if ip := getPublicIPFromMetadata(); ip != "" {
		t.Fatalf("expected no public ip, got %q", ip)
	}
}
