package wireguard

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	localstorage "github.com/in4it/wireguard-server/pkg/storage/local"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
	dateutils "github.com/in4it/wireguard-server/pkg/utils/date"
)

func TestParsePacket(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &ClientCache{
		Addresses: []ClientCacheAddresses{
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.2"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-4",
			},
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.3"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-5",
			},
		},
	}
	input := []string{
		// DNS reqs
		"45000037e04900004011cdab0abdb8020a000002e60d00350023d6861e1501000001000000000000056170706c6503636f6d0000010001",
		"4500004092d1000040111b1b0abdb8020a000002c73b0035002c4223b28e01000001000000000000037777770a676f6f676c656170697303636f6d0000410001",
		"450000e300004000fe11af480a0000020abdb8020035dbb500cffccbad65818000010000000100000975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00001c00010975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00000600010000014b004b076e732d3136333709617773646e732d313202636f02756b0011617773646e732d686f73746d617374657206616d617a6f6e03636f6d000000000100001c20000003840012750000015180",
		"450000a100004000fe11af8a0a0000020abdb8020035e136008db8bd155f81830001000000010000026462075f646e732d7364045f756470086174746c6f63616c036e657400000c0001c01c00060001000003c0004b046f726375026f72026272026e7007656c732d676d7303617474c0250d726d2d686f73746d617374657203656d730361747403636f6d0000000001000151800000271000093a8000015180",
		// http req (SYN + Data)
		"450000400000400040066ced0abdb8020a00010cc7b000507216cbdd00000000b0c2ffff008f000002040564010303060101080a69fbf8410000000004020000",
		"450200810000400040066caa0abdb8020a00010cc7b000507216cbde4845afad80180804449900000101080a69fbf873eddf46d7474554202f6c6f67696e20485454502f312e310d0a486f73743a2031302e302e312e31320d0a557365722d4167656e743a206375726c2f382e372e310d0a4163636570743a202a2f2a0d0a0d0a",
		// https req
		"450000400000400040066ced0abdb8020a00010cf24a01bb510f111000000000b0c2ffffe119000002040564010303060101080a327dff040000000004020000",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f1111c4b4fb4b801008046b8700000101080a327dff34edeeff9e",
		"4502017d0000400040066bae0abdb8020a00010cf24a01bb510f1111c4b4fb4b801808041b1500000101080a327dff36edeeff9e1603010144010001400303e3b233de9dcd3f71f4c6e3d0d45ec25144e2fcdf8c676e52ff5cfc021123786020056eefe25e5b4e9abec2953b5fa9bc1f68dd09d7ad4ddce858476b4aaaa029b80062130313021301cca9cca8ccaac030c02cc028c024c014c00a009f006b0039ff8500c400880081009d003d003500c00084c02fc02bc027c023c013c009009e0067003300be0045009c003c002f00ba0041c011c00700050004c012c0080016000a00ff01000095002b0009080304030303020301003300260024001d0020dc2b5e4f0741b2ff9982fe2bfa6641e22fe80e5b50811780b82aafae96570c2400000018001600001376706e2d7365727665722e696e3469742e696f000b00020100000a000a0008001d001700180019000d00180016080606010603080505010503080404010403020102030010000e000c02683208687474702f312e31",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f125ac4b500a3801007ee649700000101080a327dff66edeeffd1",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f125ac4b504a1801007f0609600000101080a327dff67edeeffd1",
		"4502003a0000400040066cf10abdb8020a00010cf24a01bb510f125ac4b504a180180800487100000101080a327dff6aedeeffd1140303000101",
		"450201050000400040066c260abdb8020a00010cf24a01bb510f1260c4b504a180180800ea1300000101080a327dffc0edef002a1703030035131e32cc93174219580748842686d43e1cbb73501f643eaa49b3b7ba50a9f0a97e19ec926f8b5b141b363067d9a31061b146010d8f17030300511611c04909f5346b580fe1a95c68b2a62389ca6ed7e2f31ddb38cb191cf0997e16b5efaa9248a213e621869d071af7339ddafaee642953538a03d89cb3896ecf6756f5fb80f1866671282da72dce691169170303003c3bd012039a27a373dd1b4e7509e0e9aaefc4cfae6adcae6f670501e2577e20c98233761878d9f64355a89aa389f56480517bada888a2625ef211cb5e",
	}
	now := time.Now()
	openFiles := make(PacketLoggerOpenFiles)
	for _, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		err = parsePacket(storage, data, clientCache, openFiles, now)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
	}

	out, err := storage.ReadFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.Format("2006-01-02")+".log"))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	if !strings.Contains(string(out), `,udp,10.189.184.2,10.0.0.2,58893,53,apple.com`) {
		t.Fatalf("unexpected output. Expected udp record")
	}
	if !strings.Contains(string(out), `,https,10.189.184.2,10.0.1.12,62026,443,vpn-server.in4it.io`) {
		t.Fatalf("unexpected output. Expected https record")
	}
}

func TestParsePacketSNI(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &ClientCache{
		Addresses: []ClientCacheAddresses{
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.2"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-4",
			},
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.3"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-5",
			},
		},
	}
	input := []string{
		`450000d100004000400682160abdb80240e9b468ec5001bb4f71ed891a93673d8018080468f400000101080a1329f7772c5410131603010098010000940301f1d62f57f05cc00fc8fb984e7fc381a26adc301ec143b9bab6d36f3f1b15c97200002ec014c00a0039ff850088008100350084c013c00900330045002f0041c011c00700050004c012c0080016000a00ff0100003d00000013001100000e7777772e676f6f676c652e636f6d000b00020100000a000a0008001d0017001800190010000e000c02683208687474702f312e31`,
	}
	now := time.Now()
	openFiles := make(PacketLoggerOpenFiles)
	for _, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		err = parsePacket(storage, data, clientCache, openFiles, now)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
	}

	out, err := storage.ReadFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.Format("2006-01-02")+".log"))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	if !strings.Contains(string(out), `,https,10.189.184.2,64.233.180.104,60496,443,www.google.com`) {
		t.Fatalf("unexpected output. Expected https record")
	}
}

func TestParsePacketOpenFiles(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &ClientCache{
		Addresses: []ClientCacheAddresses{
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.2"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-4",
			},
			{
				Address: net.IPNet{
					IP:   net.ParseIP("10.189.184.3"),
					Mask: net.IPMask(net.ParseIP("255.255.255.255").To4()),
				},
				ClientID: "1-2-3-5",
			},
		},
	}
	input := []string{
		// DNS reqs
		"45000037e04900004011cdab0abdb8030a000002e60d00350023d6861e1501000001000000000000056170706c6503636f6d0000010001",
		"4500004092d1000040111b1b0abdb8030a000002c73b0035002c4223b28e01000001000000000000037777770a676f6f676c656170697303636f6d0000410001",
		"450000e300004000fe11af480a0000030abdb8020035dbb500cffccbad65818000010000000100000975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00001c00010975732d656173742d310470726f6402707209616e616c797469637307636f6e736f6c65036177730361327a03636f6d00000600010000014b004b076e732d3136333709617773646e732d313202636f02756b0011617773646e732d686f73746d617374657206616d617a6f6e03636f6d000000000100001c20000003840012750000015180",
		"450000a100004000fe11af8a0a0000030abdb8020035e136008db8bd155f81830001000000010000026462075f646e732d7364045f756470086174746c6f63616c036e657400000c0001c01c00060001000003c0004b046f726375026f72026272026e7007656c732d676d7303617474c0250d726d2d686f73746d617374657203656d730361747403636f6d0000000001000151800000271000093a8000015180",
		// https reqs
		"450000400000400040066ced0abdb8020a00010cf24a01bb510f111000000000b0c2ffffe119000002040564010303060101080a327dff040000000004020000",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f1111c4b4fb4b801008046b8700000101080a327dff34edeeff9e",
		"4502017d0000400040066bae0abdb8020a00010cf24a01bb510f1111c4b4fb4b801808041b1500000101080a327dff36edeeff9e1603010144010001400303e3b233de9dcd3f71f4c6e3d0d45ec25144e2fcdf8c676e52ff5cfc021123786020056eefe25e5b4e9abec2953b5fa9bc1f68dd09d7ad4ddce858476b4aaaa029b80062130313021301cca9cca8ccaac030c02cc028c024c014c00a009f006b0039ff8500c400880081009d003d003500c00084c02fc02bc027c023c013c009009e0067003300be0045009c003c002f00ba0041c011c00700050004c012c0080016000a00ff01000095002b0009080304030303020301003300260024001d0020dc2b5e4f0741b2ff9982fe2bfa6641e22fe80e5b50811780b82aafae96570c2400000018001600001376706e2d7365727665722e696e3469742e696f000b00020100000a000a0008001d001700180019000d00180016080606010603080505010503080404010403020102030010000e000c02683208687474702f312e31",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f125ac4b500a3801007ee649700000101080a327dff66edeeffd1",
		"450000340000400040066cf90abdb8020a00010cf24a01bb510f125ac4b504a1801007f0609600000101080a327dff67edeeffd1",
		"4502003a0000400040066cf10abdb8020a00010cf24a01bb510f125ac4b504a180180800487100000101080a327dff6aedeeffd1140303000101",
		"450201050000400040066c260abdb8020a00010cf24a01bb510f1260c4b504a180180800ea1300000101080a327dffc0edef002a1703030035131e32cc93174219580748842686d43e1cbb73501f643eaa49b3b7ba50a9f0a97e19ec926f8b5b141b363067d9a31061b146010d8f17030300511611c04909f5346b580fe1a95c68b2a62389ca6ed7e2f31ddb38cb191cf0997e16b5efaa9248a213e621869d071af7339ddafaee642953538a03d89cb3896ecf6756f5fb80f1866671282da72dce691169170303003c3bd012039a27a373dd1b4e7509e0e9aaefc4cfae6adcae6f670501e2577e20c98233761878d9f64355a89aa389f56480517bada888a2625ef211cb5e",
	}
	now := time.Now()
	nowMinusOneDay := now.AddDate(0, 0, -1)
	openFiles := make(PacketLoggerOpenFiles)
	for _, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		err = parsePacket(storage, data, clientCache, openFiles, nowMinusOneDay)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
	}

	out1, err := storage.ReadFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-4-"+nowMinusOneDay.Format("2006-01-02")+".log"))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	out2, err := storage.ReadFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-5-"+nowMinusOneDay.Format("2006-01-02")+".log"))
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	if !strings.Contains(string(out2), `,udp,10.189.184.3,10.0.0.2,58893,53,apple.com`) {
		t.Fatalf("unexpected output. Expected udp record")
	}
	if !strings.Contains(string(out1), `,https,10.189.184.2,10.0.1.12,62026,443,vpn-server.in4it.io`) {
		t.Fatalf("unexpected output. Expected https record")
	}
	if len(openFiles) != 2 {
		t.Fatalf("unexpected open files count: %d", len(openFiles))
	}
	for _, s := range input {
		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		err = parsePacket(storage, data, clientCache, openFiles, now)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
	}

	dir, err := storage.ReadDir(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR))
	if err != nil {
		t.Fatalf("read dir error: %s", err)
	}
	if len(dir) != 4 {
		t.Fatalf("expected 4 files written")
	}
	if len(openFiles) != 2 {
		t.Fatalf("unexpected open files count: %d", len(openFiles))
	}
}

func TestParseTLSExtensionSNI(t *testing.T) {
	input := []string{
		"00000013001100000e7777772e676f6f676c652e636f6d000b00020100000a000a0008001d0017001800190010000e000c02683208687474702f312e31",
		"00000018001600001376706e2d7365727665722e696e3469742e696f",
		"00000018001600001376706e2d7365727665722e696e3469742e696f000b00020100000a000a0008001d001700180019000d00180016080606010603080505010503080404010403020102030010000e000c02683208687474702f312e31",
	}
	match := []string{
		"www.google.com",
		"vpn-server.in4it.io",
		"vpn-server.in4it.io",
	}
	for k, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		if sni := parseTLSExtensionSNI(data); sni != nil {
			if string(sni) != match[k] {
				t.Fatalf("got SNI, but expected different hostname. Got: %s", sni)
			}
		} else {
			t.Fatalf("no SNI found")
		}
	}
}
func TestParseTLSExtensionSNINoMatch(t *testing.T) {
	input := []string{
		"0010000e000c02",
		"000d00180016080606010603080505010503080404010403020102030010000e000c02683208687474702f312e31",
		"00",
	}
	for _, s := range input {

		data, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("hex decode error: %s", err)
		}
		if sni := parseTLSExtensionSNI(data); sni != nil {
			t.Fatalf("got match, expected no match. Got: %s", sni)
		}
	}
}

func TestCheckDiskSpace(t *testing.T) {
	err := checkDiskSpace()
	if err != nil {
		t.Fatalf("disk space error: %s", err)
	}
}

func TestPacketLoggerLogRotation(t *testing.T) {
	prefix := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)
	key1 := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s.log", time.Now().AddDate(0, 0, -1).Format("2006-01-02")))
	value1 := []byte(time.Now().Format(TIMESTAMP_FORMAT) + `,https,10.189.184.2,64.233.180.104,60496,443,www.google.com`)
	key2 := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s.log", time.Now().Format("2006-01-02")))
	value2 := []byte(time.Now().Format(TIMESTAMP_FORMAT) + `,https,10.189.184.3,64.233.180.104,12345,443,www.google.com`)

	storage := &memorystorage.MockMemoryStorage{
		Data: map[string]*memorystorage.MockReadWriterData{},
	}
	err := storage.WriteFile(key1, value1)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}
	err = storage.WriteFile(key2, value2)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}

	err = packetLoggerLogRotation(storage)
	if err != nil {
		t.Fatalf("packetLoggerRotation error: %s", err)
	}
	body, err := storage.ReadFile(key1 + ".gz")
	if err != nil {
		t.Fatalf("can't read compressed file")
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("can't open gzip reader")
	}
	bodyDecoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("can't read gzip data")
	}
	if string(bodyDecoded) != string(value1) {
		t.Fatalf("unexpected output. Got %s, expected: %s", bodyDecoded, value1)
	}
}

func TestPacketLoggerLogRotationLocalStorage(t *testing.T) {
	prefix := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)
	key1 := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s.log", time.Now().AddDate(0, 0, -1).Format("2006-01-02")))
	value1 := []byte(time.Now().Format(TIMESTAMP_FORMAT) + `,https,10.189.184.2,64.233.180.104,60496,443,www.google.com`)
	key2 := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s.log", time.Now().Format("2006-01-02")))
	value2 := []byte(time.Now().Format(TIMESTAMP_FORMAT) + `,https,10.189.184.3,64.233.180.104,12345,443,www.google.com`)

	pwd, err := os.Executable()
	if err != nil {
		t.Fatalf("os Executable error: %s", err)
	}
	storage, err := localstorage.NewWithPath(path.Dir(pwd))
	if err != nil {
		t.Fatalf("localstorage error: %s", err)
	}
	err = storage.EnsurePath(VPN_STATS_DIR)
	if err != nil {
		t.Fatalf("could not ensure path: %s", err)
	}
	storage.EnsurePath(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR))
	if err != nil {
		t.Fatalf("could not ensure path: %s", err)
	}
	err = storage.WriteFile(key1, value1)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}
	err = storage.WriteFile(key2, value2)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}
	t.Cleanup(func() {
		os.Remove(path.Join(path.Dir(pwd), key1))
		os.Remove(path.Join(path.Dir(pwd), key1+".gz.tmp"))
		os.Remove(path.Join(path.Dir(pwd), key1+".gz"))
		os.Remove(path.Join(path.Dir(pwd), key2))
	})

	err = packetLoggerLogRotation(storage)
	if err != nil {
		t.Fatalf("packetLoggerRotation error: %s", err)
	}
	body, err := storage.ReadFile(key1 + ".gz")
	if err != nil {
		t.Fatalf("can't read compressed file")
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("can't open gzip reader")
	}
	bodyDecoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("can't read gzip data")
	}
	if string(bodyDecoded) != string(value1) {
		t.Fatalf("unexpected output. Got %s, expected: %s", bodyDecoded, value1)
	}
}

func TestGetTimeUntilTomorrowStartOfDay(t *testing.T) {
	duration := getTimeUntilTomorrowStartOfDay()
	if !dateutils.DateEqual(time.Now().Add(duration), time.Now().AddDate(0, 0, 1)) {
		t.Fatalf("date is not tomorrow")
	}
}

func TestPacketLoggerLogRotationDeletion(t *testing.T) {
	prefix := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)

	storage := &memorystorage.MockMemoryStorage{
		Data: map[string]*memorystorage.MockReadWriterData{},
	}
	for i := 0; i < 20; i++ {
		timestamp := time.Now().AddDate(0, 0, -1*i)
		suffix := ".log"
		if i > 1 {
			suffix = ".log.gz"
		}
		key1 := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s%s", timestamp.Format("2006-01-02"), suffix))
		value1 := []byte(timestamp.Format(TIMESTAMP_FORMAT) + `,https,10.189.184.2,64.233.180.104,60496,443,www.google.com`)
		err := storage.WriteFile(key1, value1)
		if err != nil {
			t.Fatalf("write file error: %s", err)
		}
	}

	before, err := storage.ReadDir(prefix)
	if err != nil {
		t.Fatalf("readdir error: %s", err)
	}

	err = packetLoggerLogRotation(storage)
	if err != nil {
		t.Fatalf("packetLoggerRotation error: %s", err)
	}

	after, err := storage.ReadDir(prefix)
	if err != nil {
		t.Fatalf("readdir error: %s", err)
	}
	if len(before) != 20 {
		t.Fatalf("expected to have written 20 files. Got: %d", len(before))
	}
	if len(after) != 7 {
		t.Fatalf("only expected 7 days of retention. Got: %d", len(after))
	}
}
