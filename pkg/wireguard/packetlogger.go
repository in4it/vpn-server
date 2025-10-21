package wireguard

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
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
	"github.com/in4it/go-devops-platform/logging"
	"github.com/in4it/go-devops-platform/storage"
	dateutils "github.com/in4it/go-devops-platform/utils/date"
	pcap "github.com/packetcap/go-pcap"
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
	err = storage.EnsurePermissions(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR), 0770|os.ModeSetgid)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("could not ensure permissions of stats path: %s. Stats disabled", err))
		return
	}

	useSyscalls := runtime.GOOS == "darwin"
	handle, err := pcap.OpenLive(VPN_INTERFACE_NAME, 1600, false, 0, useSyscalls)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("can't start packet inspector: %s", err))
		return
	}
	defer handle.Close()
	i := 0
	openFiles := make(PacketLoggerOpenFiles)
	for {
		err := readPacket(storage, handle, clientCache, openFiles, vpnConfig.PacketLogsTypes)
		if err != nil {
			logging.DebugLog(fmt.Errorf("readPacket error: %s", err))
		}
		if !vpnConfig.EnablePacketLogs {
			logging.InfoLog("disabling packetlogs")
			for _, openFile := range openFiles {
				openFile.Close() //nolint:errcheck
			}
			return
		}
		if i%1000 == 0 {
			if err := checkDiskSpace(); err != nil {
				logging.ErrorLog(fmt.Errorf("disk space error: %s", err))
				for _, openFile := range openFiles {
					openFile.Close() //nolint:errcheck
				}
				return
			}
			i = 0
		}
		i++
	}
}
func readPacket(storage storage.Iface, handle *pcap.Handle, clientCache *ClientCache, openFiles PacketLoggerOpenFiles, packetLogsTypes map[string]bool) error {
	data, _, err := handle.ReadPacketData()
	if err != nil {
		return fmt.Errorf("read packet error: %s", err)
	}
	return parsePacket(storage, data, clientCache, openFiles, packetLogsTypes, time.Now())
}
func parsePacket(storage storage.Iface, data []byte, clientCache *ClientCache, openFiles PacketLoggerOpenFiles, packetLogsTypes map[string]bool, now time.Time) error {
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

	// handle open files
	logWriter, isFileOpen := openFiles[clientID+"-"+now.Format("2006-01-02")]
	if !isFileOpen {
		var err error
		filename := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, clientID+"-"+now.Format("2006-01-02")+".log")
		// check if we need to close an older writer
		for openFileKey, logWriterToClose := range openFiles {
			filenameSplit := strings.Split(openFileKey, "-")
			if len(filenameSplit) > 3 {
				dateParsed, err := time.Parse("2006-01-02", strings.Join(filenameSplit[len(filenameSplit)-3:], "-"))
				if err != nil {
					logging.ErrorLog(fmt.Errorf("packetlogger: closing unknown open file %s (cannot parse date)", filename))
					logWriterToClose.Close() //nolint:errcheck
					delete(openFiles, openFileKey)
				} else {
					if !dateutils.DateEqual(dateParsed, now) {
						logWriterToClose.Close() //nolint:errcheck
						delete(openFiles, openFileKey)
					}
				}
			} else {
				logging.ErrorLog(fmt.Errorf("packetlogger: closing file without a date %s", filename))
				logWriterToClose.Close() //nolint:errcheck
				delete(openFiles, openFileKey)
			}
		}
		// open new file for appending
		logWriter, err = storage.OpenFileForAppending(filename)
		if err != nil {
			return fmt.Errorf("could not open file for appending (%s): %s", clientID+"-"+now.Format("2006-01-02"), err)
		}
		err = storage.EnsurePermissions(filename, 0640)
		if err != nil {
			return fmt.Errorf("could not set permissions (%s): %s", clientID+"-"+now.Format("2006-01-02"), err)
		}
		openFiles[clientID+"-"+now.Format("2006-01-02")] = logWriter
	}

	logTcpVal, logTCP := packetLogsTypes["tcp"]
	logHttpVal, logHttp := packetLogsTypes["http+https"]
	logDnsVal, logDns := packetLogsTypes["dns"]
	if logTCP && !logTcpVal {
		logTCP = false
	}
	if logHttp && !logHttpVal {
		logHttp = false
	}
	if logDns && !logDnsVal {
		logDns = false
	}

	if logTCP || logHttp {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcpPacket, _ := tcpLayer.(*layers.TCP)
			if tcpPacket.SYN && logTCP {
				_, err := logWriter.Write([]byte(strings.Join([]string{
					now.Format(TIMESTAMP_FORMAT),
					"tcp",
					srcIP.String(),
					dstIP.String(),
					strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
					strconv.FormatUint(uint64(tcpPacket.DstPort), 10)},
					",") + "\n",
				))
				if err != nil {
					return fmt.Errorf("could not write to log: %s", err)
				}
			}
			if logHttp {
				switch tcpPacket.DstPort {
				case 80:
					if tcpPacket.DstPort == 80 {
						appLayer := packet.ApplicationLayer()
						if appLayer != nil {
							req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(appLayer.Payload())))
							if err != nil {
								fmt.Printf("debug: can't parse http packet: %s", err)
							} else {
								_, err = logWriter.Write([]byte(strings.Join([]string{
									now.Format(TIMESTAMP_FORMAT),
									"http",
									srcIP.String(),
									dstIP.String(),
									strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
									strconv.FormatUint(uint64(tcpPacket.DstPort), 10),
									"http://" + req.Host + req.URL.RequestURI()},
									",") + "\n",
								))
								if err != nil {
									return fmt.Errorf("could not write to log: %s", err)
								}
							}
						}
					}
				case 443:
					if tls, ok := packet.Layer(layers.LayerTypeTLS).(*layers.TLS); ok {
						for _, handshake := range tls.Handshake {
							if sni := parseTLSExtensionSNI([]byte(handshake.ClientHello.Extensions)); sni != nil {
								_, err := logWriter.Write([]byte(strings.Join([]string{
									now.Format(TIMESTAMP_FORMAT),
									"https",
									srcIP.String(),
									dstIP.String(),
									strconv.FormatUint(uint64(tcpPacket.SrcPort), 10),
									strconv.FormatUint(uint64(tcpPacket.DstPort), 10),
									string(sni)},
									",") + "\n",
								))
								if err != nil {
									return fmt.Errorf("could not write to log: %s", err)
								}
							}
						}
					}
				}
			}
		}
	}
	if logDns {
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
					_, err := logWriter.Write([]byte(strings.Join([]string{
						now.Format(TIMESTAMP_FORMAT),
						"udp",
						srcIP.String(),
						dstIP.String(),
						strconv.FormatUint(uint64(udp.SrcPort), 10),
						strconv.FormatUint(uint64(udp.DstPort), 10),
						strings.Join(questions, "#")},
						",") + "\n"))
					if err != nil {
						return fmt.Errorf("could not write to log: %s", err)
					}
				}
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
	err = unix.Statfs(wd, &stat)
	if err != nil {
		return fmt.Errorf("could not get stats from file: %s", err)
	}
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
	for {
		time.Sleep(getTimeUntilTomorrowStartOfDay()) // sleep until tomorrow
		err := packetLoggerLogRotation(storage)
		if err != nil {
			logging.ErrorLog(fmt.Errorf("packet logger log rotation error: %s", err))
		}
		err = packetLoggerRemoveTmpFiles(storage)
		if err != nil {
			logging.ErrorLog(fmt.Errorf("packet logger remove tmp files error: %s", err))
		}
	}
}

func getTimeUntilTomorrowStartOfDay() time.Duration {
	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrowStartOfDay := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 5, 0, 0, time.Local)
	return time.Until(tomorrowStartOfDay)
}

func packetLoggerLogRotation(storage storage.Iface) error {
	logDir := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)
	files, err := storage.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("readDir error: %s", err)
	}
	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return fmt.Errorf("cannot get vpn config: %s", err)
	}
	packetLogRetention := 7 // default packet log retention
	if vpnConfig.PacketLogsRetention > 0 {
		packetLogRetention = vpnConfig.PacketLogsRetention
	}
	for _, filename := range files {
		filenameWithoutSuffix := filename
		filenameWithoutSuffix = strings.TrimSuffix(filenameWithoutSuffix, ".log.gz")
		filenameWithoutSuffix = strings.TrimSuffix(filenameWithoutSuffix, ".log")
		filenameSplit := strings.Split(filenameWithoutSuffix, "-")
		if len(filenameSplit) > 3 {
			dateParsed, err := time.Parse("2006-01-02", strings.Join(filenameSplit[len(filenameSplit)-3:], "-"))
			if err == nil {
				if !dateutils.DateEqual(dateParsed, time.Now()) {
					if strings.HasSuffix(filename, ".log") {
						err := packetLoggerCompressLog(storage, filename)
						if err != nil {
							return fmt.Errorf("rotate log error: %s", err)
						}
						err = packetLoggerRenameLog(storage, filename)
						if err != nil {
							return fmt.Errorf("rotate log error (rename): %s", err)
						}
					}
					if strings.HasSuffix(filename, ".log.gz") {
						err = removeLogsAfterRetentionPeriod(storage, filename, dateParsed, packetLogRetention)
						if err != nil {
							return fmt.Errorf("remove log error (tried to remove logs after retention period has lapsed): %s", err)
						}
					}
				}

			}
		}
	}
	return nil
}

func packetLoggerRemoveTmpFiles(storage storage.Iface) error {
	files, err := storage.ReadDir(VPN_PACKETLOGGER_TMP_DIR)
	if err != nil {
		return fmt.Errorf("readDir error: %s", err)
	}
	for _, filename := range files {
		if strings.HasSuffix(filename, ".log") {
			fileInfo, err := storage.FileInfo(path.Join(VPN_PACKETLOGGER_TMP_DIR, filename))
			if err != nil {
				return fmt.Errorf("file info error (%s): %s", filename, err)
			}
			if time.Since(fileInfo.ModTime()) > (24 * time.Hour) {
				err = storage.Remove(path.Join(VPN_PACKETLOGGER_TMP_DIR, filename))
				if err != nil {
					return fmt.Errorf("file remove error (%s): %s", filename, err)
				}
			}
		}
	}
	return nil
}

func packetLoggerCompressLog(storage storage.Iface, filename string) error {
	reader, err := storage.OpenFile(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename))
	if err != nil {
		return fmt.Errorf("open file error (%s): %s", filename, err)
	}
	writer, err := storage.OpenFileForWriting(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename+".gz.tmp"))
	if err != nil {
		return fmt.Errorf("write file error (%s): %s", filename+".gz.tmp", err)
	}
	defer reader.Close() //nolint:errcheck
	defer writer.Close() //nolint:errcheck

	gzipWriter, err := gzip.NewWriterLevel(writer, gzip.DefaultCompression)
	if err != nil {
		return fmt.Errorf("gzip writer error: %s", err)
	}
	_, err = io.Copy(gzipWriter, reader)
	if err != nil {
		return fmt.Errorf("copy error: %s", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return fmt.Errorf("file close error (gzip): %s", err)
	}
	return nil
}
func packetLoggerRenameLog(storage storage.Iface, filename string) error {
	err := storage.Rename(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename+".gz.tmp"), path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename+".gz"))
	if err != nil {
		return fmt.Errorf("rename error: %s", err)
	}
	err = storage.Remove(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename))
	if err != nil {
		return fmt.Errorf("delete log error: %s", err)
	}
	return nil
}
func removeLogsAfterRetentionPeriod(storage storage.Iface, filename string, filenameDate time.Time, retentionDays int) error {
	if time.Since(filenameDate) >= (time.Duration(retentionDays) * 24 * time.Hour) {
		err := storage.Remove(path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, filename))
		if err != nil {
			return fmt.Errorf("cannot remove %s: %s", filename, err)
		}
	}
	return nil
}
