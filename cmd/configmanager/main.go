package main

import "github.com/in4it/wireguard-server/pkg/configmanager"

func main() {
	configmanager.StartServer(8081)
}
