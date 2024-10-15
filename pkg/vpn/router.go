package vpn

import (
	"net/http"

	"github.com/in4it/go-devops-platform/rest"
)

func (v *VPN) GetRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/api/vpn/connections", http.HandlerFunc(v.connectionsHandler))
	mux.Handle("/api/vpn/connection/{id}", http.HandlerFunc(v.connectionsElementHandler))
	mux.Handle("/api/vpn/connectionlicense", http.HandlerFunc(v.connectionLicenseHandler))

	mux.Handle("/api/vpn/stats/user/{date}", rest.IsAdminMiddleware(http.HandlerFunc(v.userStatsHandler)))
	mux.Handle("/api/vpn/stats/packetlogs/{user}/{date}", rest.IsAdminMiddleware(http.HandlerFunc(v.packetLogsHandler)))

	mux.Handle("/api/vpn/setup/vpn", rest.IsAdminMiddleware(http.HandlerFunc(v.vpnSetupHandler)))
	mux.Handle("/api/vpn/setup/templates", rest.IsAdminMiddleware(http.HandlerFunc(v.templateSetupHandler)))
	mux.Handle("/api/vpn/setup/restart-vpn", rest.IsAdminMiddleware(http.HandlerFunc(v.restartVPNHandler)))

	mux.Handle("/api/vpn/version", http.HandlerFunc(v.version))

	return mux
}
