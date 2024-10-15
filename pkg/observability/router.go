package observability

import "net/http"

func (o *Observability) GetRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/api/observability/", http.HandlerFunc(o.observabilityHandler))
	mux.Handle("/api/observability/ingestion/json", http.HandlerFunc(o.ingestionHandler))
	mux.Handle("/api/observability/logs", http.HandlerFunc(o.logsHandler))

	return mux
}
