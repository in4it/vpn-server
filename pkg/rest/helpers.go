package rest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Context) returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	sendCorsHeaders(w, "", c.Hostname, c.Protocol)
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`))
}

func (c *Context) write(w http.ResponseWriter, res []byte) {
	sendCorsHeaders(w, "", c.Hostname, c.Protocol)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
func (c *Context) writeWithStatus(w http.ResponseWriter, res []byte, status int) {
	sendCorsHeaders(w, "", c.Hostname, c.Protocol)
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

func isAlphaNumeric(str string) bool {
	for _, c := range str {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func getKidFromToken(token string) (string, error) {
	jwtSplit := strings.Split(token, ".")
	if len(jwtSplit) < 1 {
		return "", fmt.Errorf("token split < 1")
	}
	data, err := base64.RawURLEncoding.DecodeString(jwtSplit[0])
	if err != nil {
		return "", fmt.Errorf("could not base64 decode data part of jwt")
	}
	var header JwtHeader
	err = json.Unmarshal(data, &header)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal jwt data")
	}
	return header.Kid, nil
}

func returnIndexOrNotFound(contents []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api") {
			w.WriteHeader(http.StatusOK)
			w.Write(contents)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 page not found\n"))
		}
	})
}
