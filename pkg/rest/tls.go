package rest

import (
	"log"
	"strings"
)

func canEnableTLS(hostname string) bool {
	hostnameSplit := strings.Split(hostname, ":")
	if hostnameSplit[0] != "localhost" {
		return true
	} else {
		log.Printf("Not enabling TLS with lets encrypt. Hostname is localhost")
	}

	return false
}
