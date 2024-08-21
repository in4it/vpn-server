package wireguard

import (
	"net"
	"testing"
)

func TestGetNextFreeIPFromLisWithList(t *testing.T) {
	startIP, _, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	nextIP, err := getNextFreeIPFromList(startIP, []string{"10.189.184.2"})
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if nextIP.String() != "10.189.184.3" {
		t.Fatalf("Wrong IP: %s", nextIP)
	}
}

func TestGetNextFreeIPFromLisWithList2(t *testing.T) {
	startIP, _, err := net.ParseCIDR("10.189.184.1/21")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	nextIP, err := getNextFreeIPFromList(startIP, []string{"10.190.190.2", "10.189.184.2", "10.190.190.3"})
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if nextIP.String() != "10.189.184.3" {
		t.Fatalf("Wrong IP: %s", nextIP)
	}
}
