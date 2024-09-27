package observability

import (
	"testing"
	"time"
)

func TestFloatToDate2Way(t *testing.T) {
	now := time.Now()
	float := DateToFloat(now)
	date := FloatToDate(float)
	if date.Format(TIMESTAMP_FORMAT) != now.Format(TIMESTAMP_FORMAT) {
		t.Fatalf("got: %s, expected: %s", date.Format(TIMESTAMP_FORMAT), now.Format(TIMESTAMP_FORMAT))
	}
}
