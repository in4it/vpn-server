package wireguard

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/packetcap/go-pcap"
)

func RunPacketLogger(storage storage.Iface, clientCache *ClientCache) {
	useSyscalls := false
	if runtime.GOOS == "darwin" {
		useSyscalls = true
	}
	handle, err := pcap.OpenLive(VPN_INTERFACE_NAME, 1600, false, 0, useSyscalls)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("can't start packet inspector: %s", err))
		return
	}
	for {
		err := readPacket(storage, handle, clientCache)
		if err != nil {
			logging.DebugLog(fmt.Errorf("readPacket error: %s", err))
		}
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
	now := time.Now()
	filename := path.Join(VPN_STATS_DIR, "ip-"+now.Format("2006-01-02.log"))
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
				//fmt.Printf("payload: %s", tcpPacket.LayerContents())
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
			//fmt.Printf("dstport: %s: %s\n", tcpPacket.DstPort.String(), packet.ApplicationLayer().Payload())
			// TODO: get SNI from handshake
			/*if tls, ok := packet.Layer(layers.LayerTypeTLS).(*layers.TLS); ok {
				fmt.Printf("handshake: %+v\n", tls.Handshake)
			}*/
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
