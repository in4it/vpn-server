package wireguard

import (
	"errors"
	"io"
	"net"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
)

func TestSanitizeLogField(t *testing.T) {
	tests := map[string]string{
		"www.google.com":                     "www.google.com",
		"http://example.com/path?a=1&b=%20x": "http://example.com/path?a=1&b=%20x",
		"evil.com,injected":                  "evil.com%2Cinjected",
		"evil.com\n2024-01-01T00:00:00,tcp":  "evil.com%0A2024-01-01T00:00:00%2Ctcp",
		"a\r\tb\x00c\x7f":                    "a%0D%09b%00c%7F",
		"q1#q2":                              "q1%23q2",
		"caf\xc3\xa9":                        "caf%C3%A9",
	}
	for input, want := range tests {
		if got := sanitizeLogField(input); got != want {
			t.Errorf("sanitizeLogField(%q) = %q, want %q", input, got, want)
		}
	}
}

func dnsQueryPacket(t *testing.T, srcIP, dstIP string, names ...string) []byte {
	t.Helper()
	ip := &layers.IPv4{Version: 4, TTL: 64, Protocol: layers.IPProtocolUDP, SrcIP: net.ParseIP(srcIP).To4(), DstIP: net.ParseIP(dstIP).To4()}
	udp := &layers.UDP{SrcPort: 51234, DstPort: 53}
	if err := udp.SetNetworkLayerForChecksum(ip); err != nil {
		t.Fatalf("checksum error: %s", err)
	}
	dns := &layers.DNS{ID: 1, RD: true, QDCount: uint16(len(names))}
	for _, name := range names {
		dns.Questions = append(dns.Questions, layers.DNSQuestion{Name: []byte(name), Type: layers.DNSTypeA, Class: layers.DNSClassIN})
	}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}, ip, udp, dns); err != nil {
		t.Fatalf("serialize error: %s", err)
	}
	return buf.Bytes()
}

// TestParsePacketDNSInjection: dns names come from client traffic and must not
// be able to add fields or lines to the packet log.
func TestParsePacketDNSInjection(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &ClientCache{}
	if err := UpdateClientCache(PeerConfig{ID: "1-2-3-4-1", Address: "10.189.184.2/32"}, clientCache); err != nil {
		t.Fatalf("UpdateClientCache error: %s", err)
	}
	now := time.Now()
	openFiles := make(PacketLoggerOpenFiles)
	data := dnsQueryPacket(t, "10.189.184.2", "10.0.0.2", "evil.com,x\n2000-01-01T00:00:00,tcp,1.1.1.1,2.2.2.2,1,2", "a#b.com")
	if err := parsePacket(storage, data, clientCache, openFiles, map[string]bool{"dns": true}, now); err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out, err := storage.ReadFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.Format("2006-01-02")+".log"))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d: %q", len(lines), out)
	}
	fields := strings.Split(lines[0], ",")
	if len(fields) != 7 {
		t.Fatalf("expected 7 fields, got %d: %q", len(fields), lines[0])
	}
	if questions := strings.Split(fields[6], "#"); len(questions) != 2 {
		t.Fatalf("expected 2 dns questions, got %d: %q", len(questions), fields[6])
	}
}

type fakeWriteCloser struct {
	closed bool
}

func (f *fakeWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (f *fakeWriteCloser) Close() error                { f.closed = true; return nil }

var _ io.WriteCloser = &fakeWriteCloser{}

func TestCloseOpenFiles(t *testing.T) {
	f1, f2 := &fakeWriteCloser{}, &fakeWriteCloser{}
	openFiles := PacketLoggerOpenFiles{"a-2024-01-01": f1, "b-2024-01-01": f2}
	closeOpenFiles(openFiles)
	if !f1.closed || !f2.closed {
		t.Fatalf("expected files to be closed")
	}
	if len(openFiles) != 0 { // closed writers must not be reused after resuming
		t.Fatalf("expected open files map to be empty, got %d", len(openFiles))
	}
}

// TestWaitForDiskSpace: the packet logger resumes once there is enough disk
// space again, instead of stopping for good.
func TestWaitForDiskSpace(t *testing.T) {
	checks := 0
	check := func() error {
		checks++
		if checks < 3 {
			return errors.New("not enough disk space")
		}
		return nil
	}
	sleeps := []time.Duration{}
	sleep := func(d time.Duration) { sleeps = append(sleeps, d) }
	if !waitForDiskSpace(check, func() bool { return true }, sleep, time.Minute) {
		t.Fatalf("expected the packet logger to resume")
	}
	if checks != 3 || len(sleeps) != 3 || sleeps[0] != time.Minute {
		t.Fatalf("unexpected checks (%d) or sleeps (%v)", checks, sleeps)
	}

	// packet logs disabled while waiting
	if waitForDiskSpace(func() error { return errors.New("full") }, func() bool { return false }, func(time.Duration) {}, time.Minute) {
		t.Fatalf("expected the packet logger to stop when packet logs are disabled")
	}
}
