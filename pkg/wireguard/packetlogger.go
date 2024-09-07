package wireguard

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	dateutils "github.com/in4it/wireguard-server/pkg/utils/date"
	"github.com/packetcap/go-pcap"
	"golang.org/x/sys/unix"
)

var (
	PacketLoggerIsRunning sync.Mutex
)

func RunPacketLogger(storage storage.Iface, clientCache *ClientCache, vpnConfig *VPNConfig) {
	if !vpnConfig.EnablePacketLogs {
		return
	}
	fmt.Printf("starting packetlogger")
	// ensure we only run a single instance of the packet logger
	PacketLoggerIsRunning.Lock()
	defer PacketLoggerIsRunning.Unlock()
	// ensure logs dir is created
	err := storage.EnsurePath(VPN_STATS_DIR)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not create stats path: %s. Stats disabled", err))
		return
	}
	err = storage.EnsureOwnership(VPN_STATS_DIR, "vpn")
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not ensure ownership of stats path: %s. Stats disabled", err))
		return
	}
	err = storage.EnsurePath(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR))
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not create stats path: %s. Stats disabled", err))
		return
	}
	err = storage.EnsureOwnership(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR), "vpn")
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not ensure ownership of stats path: %s. Stats disabled", err))
		return
	}

	useSyscalls := false
	if runtime.GOOS == "darwin" {
		useSyscalls = true
	}
	handle, err := pcap.OpenLive(VPN_INTERFACE_NAME, 1600, false, 0, useSyscalls)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("can't start packet inspector: %s", err))
		return
	}
	defer handle.Close()
	i := 0
	for {
		err := readPacket(storage, handle, clientCache)
		if err != nil {
			logging.DebugLog(fmt.Errorf("readPacket error: %s", err))
		}
		if !vpnConfig.EnablePacketLogs {
			logging.InfoLog("disabling packetlogs")
			return
		}
		if i%1000 == 0 {
			if err := checkDiskSpace(); err != nil {
				logging.ErrorLog(fmt.Errorf("disk space error: %s", err))
				return
			}
			i = 0
		}
		i++
	}
}
func readPacket(storage storage.Iface, handle *pcap.Handle, clientCache *ClientCache) error {
	data, _, err := handle.ReadPacketData()
	if err != nil {
		return fmt.Errorf("read packet error: %s", err)
	}
	return parsePacket(storage, data, clientCache)
}
func parsePacket(storage storage.Iface, data []byte, clientCache *ClientCache) error {
	packet := gopacket.NewPacket(data, layers.IPProtocolIPv4, gopacket.DecodeOptions{Lazy: true, DecodeStreamsAsDatagrams: true})
	var (
		ip4   *layers.IPv4
		ip6   *layers.IPv6
		srcIP net.IP
		dstIP net.IP
	)

	if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ip4 = ipv4Layer.(*layers.IPv4)
		srcIP = ip4.SrcIP
		dstIP = ip4.DstIP
	}
	if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
		ip6 = ipv6Layer.(*layers.IPv6)
		srcIP = ip6.SrcIP
		dstIP = ip6.DstIP
	}
	if ip4 == nil && ip6 == nil {
		return fmt.Errorf("got packet which is not ipv4/ipv6")
	}

	clientID := ""
	for _, address := range clientCache.Addresses {
		if address.Address.Contains(srcIP) {
			clientID = address.ClientID
		}
	}
	if clientID == "" { // doesn't match a client ID
		return nil
	}
	now := time.Now()
	filename := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, clientID+"-"+now.Format("2006-01-02")+".log")
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcpPacket, _ := tcpLayer.(*layers.TCP)
		if tcpPacket.SYN {
			storage.AppendFile(filename, []byte(strings.Join([]string{
				time.Now().Format(TIMESTAMP_FORMAT),
				"tcp",
				srcIP.String(),
				dstIP.String(),
				strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
				strconv.FormatUint(uint64(tcpPacket.DstPort), 10)},
				",")+"\n",
			))
		}
		switch tcpPacket.DstPort {
		case 80:
			if tcpPacket.DstPort == 80 {
				appLayer := packet.ApplicationLayer()
				if appLayer != nil {
					req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(appLayer.Payload())))
					if err != nil {
						fmt.Printf("debug: can't parse http packet: %s", err)
					} else {
						storage.AppendFile(filename, []byte(strings.Join([]string{
							time.Now().Format(TIMESTAMP_FORMAT),
							"http",
							srcIP.String(),
							dstIP.String(),
							strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
							strconv.FormatUint(uint64(tcpPacket.DstPort), 10),
							"http://" + req.Host + req.URL.RequestURI()},
							",")+"\n",
						))
					}
				}
			}
		case 443:
			if tls, ok := packet.Layer(layers.LayerTypeTLS).(*layers.TLS); ok {
				for _, handshake := range tls.Handshake {
					if sni := parseTLSExtensionSNI([]byte(handshake.ClientHello.Extensions)); sni != nil {
						storage.AppendFile(filename, []byte(strings.Join([]string{
							time.Now().Format(TIMESTAMP_FORMAT),
							"https",
							srcIP.String(),
							dstIP.String(),
							strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
							strconv.FormatUint(uint64(tcpPacket.DstPort), 10),
							string(sni)},
							",")+"\n",
						))
					}
				}
			}
		}
	}
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)

		if udp.NextLayerType().Contains(layers.LayerTypeDNS) {
			dnsPacket := packet.Layer(layers.LayerTypeDNS)
			if dnsPacket != nil {
				udpDNS := dnsPacket.(*layers.DNS)
				questions := []string{}
				for k := range udpDNS.Questions {
					found := false
					for _, question := range questions {
						if question == string(udpDNS.Questions[k].Name) {
							found = true
						}
					}
					if !found {
						questions = append(questions, string(udpDNS.Questions[k].Name))
					}

				}
				storage.AppendFile(filename, []byte(strings.Join([]string{
					time.Now().Format(TIMESTAMP_FORMAT),
					"udp",
					srcIP.String(),
					dstIP.String(),
					strconv.FormatUint(uint64(udp.SrcPort), 10),
					strconv.FormatUint(uint64(udp.DstPort), 10),
					strings.Join(questions, "#")},
					",")+"\n"))
			}
		}
	}

	return nil
}

// TLS Extensions http://www.iana.org/assignments/tls-extensiontype-values/tls-extensiontype-values.xhtml
type TLSExtension uint16

const (
	ExtServerName           TLSExtension = 0
	ExtMaxFragLen           TLSExtension = 1
	ExtClientCertURL        TLSExtension = 2
	ExtTrustedCAKeys        TLSExtension = 3
	ExtTruncatedHMAC        TLSExtension = 4
	ExtStatusRequest        TLSExtension = 5
	ExtUserMapping          TLSExtension = 6
	ExtClientAuthz          TLSExtension = 7
	ExtServerAuthz          TLSExtension = 8
	ExtCertType             TLSExtension = 9
	ExtSupportedGroups      TLSExtension = 10
	ExtECPointFormats       TLSExtension = 11
	ExtSRP                  TLSExtension = 12
	ExtSignatureAlgs        TLSExtension = 13
	ExtUseSRTP              TLSExtension = 14
	ExtHeartbeat            TLSExtension = 15
	ExtALPN                 TLSExtension = 16
	ExtStatusRequestV2      TLSExtension = 17
	ExtSignedCertTS         TLSExtension = 18
	ExtClientCertType       TLSExtension = 19
	ExtServerCertType       TLSExtension = 20
	ExtPadding              TLSExtension = 21
	ExtEncryptThenMAC       TLSExtension = 22
	ExtExtendedMasterSecret TLSExtension = 23
	ExtSessionTicket        TLSExtension = 35
	ExtNPN                  TLSExtension = 13172
	ExtRenegotiationInfo    TLSExtension = 65281
)

func parseTLSExtensionSNI(data []byte) []byte {
	for len(data) > 0 {
		if len(data) < 4 {
			break
		}
		extensionType := binary.BigEndian.Uint16(data[:2])
		length := binary.BigEndian.Uint16(data[2:4])
		if len(data) < 4+int(length) {
			break
		}
		if TLSExtension(extensionType) == ExtServerName && len(data) > 6 {
			serverNameExtensionLength := binary.BigEndian.Uint16(data[4:6])
			entryType := data[6]

			if serverNameExtensionLength > 0 && entryType == 0 && len(data) > 8 { // 0 = DNS hostname
				hostnameLength := binary.BigEndian.Uint16(data[7:9])
				if len(data) > int(8+hostnameLength) {
					return data[9 : 9+hostnameLength]
				}
			}
		}
		data = data[4+length:]
	}
	return nil
}

func checkDiskSpace() error {
	var stat unix.Statfs_t

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("cannot get cwd: %s", err)
	}
	unix.Statfs(wd, &stat)
	if stat.Blocks*uint64(stat.Bsize) == 0 {
		return fmt.Errorf("no blocks available")
	}
	freeDiskSpace := float64(stat.Bfree) / float64(stat.Blocks)
	if freeDiskSpace < 0.10 {
		return fmt.Errorf("not enough disk free disk space: %f", freeDiskSpace)
	}

	return nil
}

// Packet log rotation
func PacketLoggerLogRotation(storage storage.Iface) {
	err := packetLoggerLogRotation(storage)
	logging.ErrorLog(fmt.Errorf("packet logger log rotation error: %s", err))
}

func packetLoggerLogRotation(storage storage.Iface) error {
	logDir := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)
	files, err := storage.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("readDir error: %s", err)
	}
	for _, filename := range files {
		filenameSplit := strings.Split(strings.TrimSuffix(filename, ".log"), "-")
		if len(filenameSplit) > 3 {
			dateParsed, err := time.Parse("2006-01-02", filenameSplit[len(filenameSplit)-3])
			if err == nil {
				if !dateutils.DateEqual(dateParsed, time.Now()) {
					err := packetLoggerRotateLog(storage, filename)
					if err != nil {
						return fmt.Errorf("rotate log error: %s", err)
					}
				}
			}
		}
	}
	return nil
}

func packetLoggerRotateLog(storage storage.Iface, filename string) error {
	reader, err := storage.OpenFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename))
	if err != nil {
		return fmt.Errorf("open file error (%s): %s", filename, err)
	}
	writer, err := storage.OpenFileForWriting(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename+".gz.tmp"))
	if err != nil {
		return fmt.Errorf("write file error (%s): %s", filename+".gz.tmp", err)
	}
	defer reader.Close()
	defer writer.Close()
	// compress, write, rename
	return nil
}
