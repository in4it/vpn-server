package configmanager

import "net/http"

func (c *ConfigManager) getRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/pubkey", http.HandlerFunc(c.getPubKey))
	mux.Handle("/refresh-clients", http.HandlerFunc(c.refreshClients))
	mux.Handle("/upgrade", http.HandlerFunc(c.upgrade))
	mux.Handle("/version", http.HandlerFunc(c.version))

	return mux
}
