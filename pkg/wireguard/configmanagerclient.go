package wireguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// variables so tests can override them
var (
	configManagerURI       = CONFIGMANAGER_URI
	configManagerTransport http.RoundTripper // nil: http.DefaultTransport
)

// notifyConfigManager posts the payload (json encoded, if not nil) to the
// configmanager and expects a 202 Accepted
func notifyConfigManager(uri string, payload any) error {
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: configManagerTransport,
	}
	var body io.Reader
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("could not marshal configmanager request: %s", err)
		}
		body = bytes.NewReader(payloadBytes)
	}
	resp, err := client.Post("http://"+configManagerURI+uri, "application/json", body)
	if err != nil {
		return fmt.Errorf("configmanager post error: %s", err)
	}
	defer resp.Body.Close()                                        //nolint:errcheck
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024)) // drain, so the connection can be reused
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
	}
	return nil
}

func notifyConfigManagerRefreshClients(action string, filenames []string) error {
	return notifyConfigManager("/refresh-clients", RefreshClientRequest{
		Action:    action,
		Filenames: filenames,
	})
}
