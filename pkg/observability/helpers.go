package observability

import (
	"fmt"
	"net/http"
	"strings"
)

func (o *Observability) returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`))
}
