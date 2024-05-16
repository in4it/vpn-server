package main

import (
	"flag"

	"github.com/in4it/wireguard-server/pkg/rest"
)

func main() {
	var (
		httpPort  int
		httpsPort int
	)
	flag.IntVar(&httpPort, "http-port", 80, "http port to run server on")
	flag.IntVar(&httpsPort, "https-port", 443, "https port to run server on")
	flag.Parse()
	rest.StartServer(httpPort, httpsPort, rest.SERVER_TYPE_VPN)
}
