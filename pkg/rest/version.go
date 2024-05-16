package rest

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

//go:generate cp -r ../../latest ./resources/version
//go:embed resources/version

var version string

func (c *Context) version(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		out, err := json.Marshal(map[string]string{"version": strings.TrimSpace(version)})
		if err != nil {
			c.returnError(w, fmt.Errorf("version marshal error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) upgrade(w http.ResponseWriter, r *http.Request) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(r.Method, "http://"+wireguard.CONFIGMANAGER_URI+"/upgrade", nil)
	if err != nil {
		c.returnError(w, fmt.Errorf("upgrade request error: %s", err), http.StatusBadRequest)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		c.returnError(w, fmt.Errorf("upgrade error: %s", err), http.StatusBadRequest)
		return
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.returnError(w, fmt.Errorf("upgrade error: got status code: %d. Respons: %s", resp.StatusCode, bodyBytes), http.StatusBadRequest)
			return
		}
		c.returnError(w, fmt.Errorf("upgrade error: got status code: %d. Couldn't get response", resp.StatusCode), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.returnError(w, fmt.Errorf("body read error: %s", err), http.StatusBadRequest)
		return
	}

	c.write(w, bodyBytes)

}
