package observability

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (o *Observability) observabilityHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (o *Observability) ingestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := o.Ingest(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("error: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (o *Observability) logsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.FormValue("fromDate") == "" {
		o.returnError(w, fmt.Errorf("no from date supplied"), http.StatusBadRequest)
		return
	}
	fromDate, err := time.Parse("2006-01-02", r.FormValue("fromDate"))
	if err != nil {
		o.returnError(w, fmt.Errorf("invalid date: %s", err), http.StatusBadRequest)
		return
	}
	if r.FormValue("endDate") == "" {
		o.returnError(w, fmt.Errorf("no end date supplied"), http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", r.FormValue("endDate"))
	if err != nil {
		o.returnError(w, fmt.Errorf("invalid date: %s", err), http.StatusBadRequest)
		return
	}
	offset := 0
	if r.FormValue("offset") != "" {
		i, err := strconv.Atoi(r.FormValue("offset"))
		if err == nil {
			offset = i
		}
	}
	maxLines := 0
	if r.FormValue("maxLines") != "" {
		i, err := strconv.Atoi(r.FormValue("maxLines"))
		if err == nil {
			maxLines = i
		}
	}
	pos := int64(0)
	if r.FormValue("pos") != "" {
		i, err := strconv.ParseInt(r.FormValue("pos"), 10, 64)
		if err == nil {
			pos = i
		}
	}
	displayTags := strings.Split(r.FormValue("display-tags"), ",")
	filterTagsSplit := strings.Split(r.FormValue("filter-tags"), ",")
	filterTags := []KeyValue{}
	for _, tag := range filterTagsSplit {
		kv := strings.Split(tag, "=")
		if len(kv) == 2 {
			filterTags = append(filterTags, KeyValue{Key: kv[0], Value: kv[1]})
		}
	}
	out, err := o.getLogs(fromDate, endDate, pos, maxLines, offset, r.FormValue("search"), displayTags, filterTags)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("get logs error: %s", err)
		return
	}
	outBytes, err := json.Marshal(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("marshal error: %s", err)
		return
	}
	w.Write(outBytes)
}
