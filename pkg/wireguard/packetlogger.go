package wireguard

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/packetcap/go-pcap"
)

func RunPacketLogger(storage storage.Iface) {
	useSyscalls := false
	if runtime.GOOS == "darwin" {
		useSyscalls = true
	}
	handle, err := pcap.OpenLive("en0" /*VPN_INTERFACE_NAME */, 1600, true, 0, useSyscalls)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("can't start packet inspector: %s", err))
	}
	for {
		err := readPacket(handle)
		if err != nil {
			logging.DebugLog(fmt.Errorf("readPacket error: %s", err))
		}
	}
}
func readPacket(handle *pcap.Handle) error {
	data, _, err := handle.ReadPacketData()
	if err != nil {
		return fmt.Errorf("read packet error: %s", err)
	}

	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Lazy)
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
				fmt.Printf("DNS Req: %s\n", strings.Join(questions, ", "))
			}
		}

	}

	return nil
}
