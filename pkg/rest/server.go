package rest

import (
	"crypto/tls"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/in4it/wireguard-server/pkg/logging"
	localstorage "github.com/in4it/wireguard-server/pkg/storage/local"
	"golang.org/x/crypto/acme/autocert"
)

var (
	//go:embed static
	assets             embed.FS
	enableTLSWaiter    chan (bool) = make(chan bool)
	TLSWaiterCompleted bool
)

func StartServer(httpPort, httpsPort int, serverType string) {
	localStorage, err := localstorage.New()
	if err != nil {
		log.Fatalf("couldn't initialize storage: %s", err)
	}
	c, err := newContext(localStorage, serverType)
	if err != nil {
		log.Fatalf("startup failed: %s", err)
	}

	go handleSignals(c)

	assetsFS, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatalf("could not load static web assets")
	}

	indexHtml, err := assetsFS.Open("index.html")
	if err != nil {
		log.Fatalf("could not load static web assets (index.html)")
	}
	indexHtmlBody, err := io.ReadAll(indexHtml)
	if err != nil {
		log.Fatalf("could not read static web assets (index.html)")
	}

	certManager := autocert.Manager{}

	// HTTP Configuration
	go func() { // start http server
		log.Printf("Start http server on port %d", httpPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), certManager.HTTPHandler(c.loggingMiddleware(c.httpsRedirectMiddleware(c.corsMiddleware(c.getRouter(assetsFS, indexHtmlBody)))))))
	}()

	// TLS Configuration
	if !c.EnableTLS || !canEnableTLS(c.Hostname) {
		<-enableTLSWaiter
	}
	// only enable when TLS is enabled

	logging.DebugLog(fmt.Errorf("enabling TLS endpoint with let's encrypt for hostname '%s'", c.Hostname))
	certManager.Prompt = autocert.AcceptTOS
	certManager.HostPolicy = autocert.HostWhitelist(c.Hostname)
	certManager.Cache = autocert.DirCache("tls-certs")
	tlsServer := &http.Server{
		Addr: fmt.Sprintf(":%d", httpsPort),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: c.loggingMiddleware(c.corsMiddleware(c.getRouter(assetsFS, indexHtmlBody))),
	}
	c.Protocol = "https"
	TLSWaiterCompleted = true
	log.Fatal(tlsServer.ListenAndServeTLS("", ""))
}
