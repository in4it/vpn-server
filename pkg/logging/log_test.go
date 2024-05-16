package logging

import "testing"

func TestLog(t *testing.T) {
	Loglevel = LOG_DEBUG
	if Loglevel&LOG_DEBUG != LOG_DEBUG {
		t.Fatalf("log level is not debugging1")
	}
	Loglevel = LOG_DEBUG + LOG_ERROR
	if Loglevel&LOG_DEBUG != LOG_DEBUG {
		t.Fatalf("log level is not debugging: %d vs %d", Loglevel&LOG_DEBUG, Loglevel)
	}
	Loglevel = LOG_ERROR
	if Loglevel&LOG_DEBUG == LOG_DEBUG {
		t.Fatalf("log level is debugging: %d vs %d", Loglevel&LOG_DEBUG, Loglevel)
	}

}
