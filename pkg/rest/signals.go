package rest

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func handleSignals(c *Context) {
	// write pid file so other process can find it
	err := os.WriteFile(path.Join(c.AppDir, "rest-server.pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	if err != nil {
		log.Printf("Could not write pid file\n")
	}
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)
	for sig := range signalChannel {
		switch sig {
		case syscall.SIGHUP:
			c.ReloadConfig()
		}
	}
}

func (c *Context) ReloadConfig() {
	newC, err := newContext(c.Storage.Client, SERVER_TYPE_VPN)
	if err != nil {
		log.Printf("ReloadConfig failed: %s\n", err)
	}
	c.AppDir = newC.AppDir
	c.Hostname = newC.Hostname
	c.SetupCompleted = newC.SetupCompleted
	c.UserStore = newC.UserStore
	log.Printf("Config Reloaded!\n")
}
