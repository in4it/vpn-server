package observability

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

func (o *Observability) returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + strings.Replace(err.Error(), `"`, `\"`, -1) + `"}`))
}

func FloatToDate(datetime float64) time.Time {
	datetimeInt := int64(datetime)
	decimals := datetime - float64(datetimeInt)
	nsecs := int64(math.Round(decimals * 1_000_000)) // precision to match golang's time.Time
	return time.Unix(datetimeInt, nsecs*1000)
}

func DateToFloat(datetime time.Time) float64 {
	seconds := float64(datetime.Unix())
	nanoseconds := float64(datetime.Nanosecond()) / 1e9
	fmt.Printf("nanosec: %f", nanoseconds)
	return seconds + nanoseconds
}
