package vpn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//go:generate cp -r ../../latest ./resources/version
//go:embed resources/version

var version string

func (v *VPN) version(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		out, err := json.Marshal(map[string]string{"version": strings.TrimSpace(version)})
		if err != nil {
			v.returnError(w, fmt.Errorf("version marshal error: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	default:
		v.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
