package observability

import (
	"fmt"
	"net/http"
)

func (o *Observability) observabilityHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (o *Observability) ingestionHandler(w http.ResponseWriter, r *http.Request) {
	msgs, err := Decode(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("Got msgs: %+v\n", msgs)
	w.WriteHeader(http.StatusOK)
}
