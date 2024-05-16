package scim

import (
	"fmt"
	"net/http"
)

func returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
}

func writeWithStatus(w http.ResponseWriter, res []byte, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
}

func write(w http.ResponseWriter, res []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
