package observability

import (
	"fmt"
	"net/http"
)

func (o *Observability) observabilityHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (o *Observability) ingestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	msgs, err := Decode(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("error: %s", err)
		return
	}
	_, err = o.Buffer.Write(encodeMessage(msgs))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("cannot store message: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
