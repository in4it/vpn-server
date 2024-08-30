package wireguard

import (
	"encoding/hex"
	"net"
	"path"
	"strings"
	"testing"
	"time"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
)

func TestParsePacket(t *testing.T) {
	storage := &testingmocks.MockMemoryStorage{}
	clientCache := &ClientCache{
		Addresses: []ClientCacheAddresses{
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.2"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-4-1",
			},
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.3"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-5-1",
			},
		},
	}
	input := []string{
		"45000037e04900004011cdab0abdb8020a000002e60d00350023d6861e1501000001000000000000056170706c6503636f6d0000010001",
		"4500004092d1000040111b1b0abdb8020a000002c73b0035002c4223b28e01000001000000000000037777770a676f6f676c656170697303636f6d0000410001",
		"450000e300004000fe11af480a0000020abdb8020035dbb500cffccbad65818000010000000100000975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00001c00010975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00000600010000014b004b076e732d3136333709617773646e732d313202636f02756b0011617773646e732d686f73746d617374657206616d617a6f6e03636f6d000000000100001c20000003840012750000015180",
		"450000a100004000fe11af8a0a0000020abdb8020035e136008db8bd155f81830001000000010000026462075f646e732d7364045f756470086174746c6f63616c036e657400000c0001c01c00060001000003c0004b046f726375026f72026272026e7007656c732d676d7303617474c0250d726d2d686f73746d617374657203656d730361747403636f6d0000000001000151800000271000093a8000015180",
	}
	now := time.Now()
	for _, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		err = parsePacket(storage, data, clientCache)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
	}

	out, err := storage.ReadFile(path.Join(VPN_STATS_DIR, "ip-"+now.Format("2006-01-02.log")))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	if !strings.Contains(string(out), `udp,10.189.184.2,10.0.0.2,58893,53(domain),apple.com`) {
		t.Fatalf("unexpected output")
	}
}
