package vpn

import (
	"fmt"
	"net/http"
	"strings"
)

func (v *VPN) returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`))
}

func (v *VPN) write(w http.ResponseWriter, res []byte) {
	sendCorsHeaders(w, "", v.Hostname, v.Protocol)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
func (v *VPN) writeWithStatus(w http.ResponseWriter, res []byte, status int) {
	sendCorsHeaders(w, "", v.Hostname, v.Protocol)
	w.WriteHeader(status)
	w.Write(res)
}

func sendCorsHeaders(w http.ResponseWriter, headers string, hostname string, protocol string) {
	if hostname == "" {
		w.Header().Add("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", protocol, hostname))
	}
	w.Header().Add("Access-Control-allow-methods", "GET,HEAD,POST,PUT,OPTIONS,DELETE,PATCH")
	if headers != "" {
		w.Header().Add("Access-Control-Allow-Headers", headers)
	}
}
