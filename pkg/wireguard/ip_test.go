package wireguard

import (
	"net"
	"strings"
	"testing"
)

func TestGetNextFreeIPFromLisWithList(t *testing.T) {
	startIP, addressRange, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	nextIP, err := getNextFreeIPFromList(startIP, addressRange, []string{"10.189.184.2"}, "/32")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if nextIP.String() != "10.189.184.3" {
		t.Fatalf("Wrong IP: %s", nextIP)
	}
}

func TestGetNextFreeIPFromLisWithList2(t *testing.T) {
	startIP, addressRange, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	nextIP, err := getNextFreeIPFromList(startIP, addressRange, []string{"10.190.190.2", "10.189.184.2", "10.190.190.3"}, "/32")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if nextIP.String() != "10.189.184.3" {
		t.Fatalf("Wrong IP: %s", nextIP)
	}
}

func TestGetNextFreeIPWithRange(t *testing.T) {
	startIP, addressRange, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	networkPrefix := []string{
		"/32",
		"/32",
		"/32",
		"/32",
		"/32",
		"/32",
		"/30",
		"/30",
		"/32",
	}
	testCases := [][]string{
		{},
		{"10.189.184.2"},
		{"10.189.184.2/32"},
		{"10.189.184.2", "10.189.184.3", "10.189.184.4/30"},
		{"10.189.184.2", "10.189.184.3", "10.189.184.4/30", "10.189.184.8/32"},
		{"10.189.184.1/30", "10.189.184.4/30", "10.189.184.8/30"},
		{},
		{"10.189.184.4/30", "10.189.184.8/30"},
		{"10.189.189.2/32", "10.189.189.3/32", "10.189.189.4/32"},
	}
	expected := []string{
		"10.189.184.2",
		"10.189.184.3",
		"10.189.184.3",
		"10.189.184.8",
		"10.189.184.9",
		"10.189.184.12",
		"10.189.184.4",
		"10.189.184.12",
		"10.189.184.2",
	}

	for k := range testCases {
		nextIP, err := getNextFreeIPFromList(startIP, addressRange, testCases[k], networkPrefix[k])
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		if nextIP.String() != expected[k] {
			t.Fatalf("Wrong IP: %s", nextIP)
		}
	}

}

func TestIPNotInRange(t *testing.T) {
	startIP, addressRange, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	_, err = getNextFreeIPFromList(startIP, addressRange, []string{"10.189.188.0/22"}, "/22")
	if !strings.Contains(err.Error(), "not within address range") {
		t.Fatalf("Expected error, got: %s", err)
	}

}
