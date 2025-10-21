package vpn

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (v *VPN) vpnSetupHandler(w http.ResponseWriter, r *http.Request) {
	vpnConfig, err := wireguard.GetVPNConfig(v.Storage)
	if err != nil {
		v.returnError(w, fmt.Errorf("could not get vpn config: %s", err), http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		packetLogTypes := []string{}
		for k, enabled := range vpnConfig.PacketLogsTypes {
			if enabled {
				packetLogTypes = append(packetLogTypes, k)
			}
		}
		if vpnConfig.PacketLogsRetention == 0 {
			vpnConfig.PacketLogsRetention = 7
		}
		setupRequest := VPNSetupRequest{
			Routes:              strings.Join(vpnConfig.ClientRoutes, ", "),
			VPNEndpoint:         vpnConfig.Endpoint,
			AddressRange:        vpnConfig.AddressRange.String(),
			ClientAddressPrefix: vpnConfig.ClientAddressPrefix,
			Port:                strconv.Itoa(vpnConfig.Port),
			ExternalInterface:   vpnConfig.ExternalInterface,
			Nameservers:         strings.Join(vpnConfig.Nameservers, ","),
			DisableNAT:          vpnConfig.DisableNAT,
			EnablePacketLogs:    vpnConfig.EnablePacketLogs,
			PacketLogsTypes:     packetLogTypes,
			PacketLogsRetention: strconv.Itoa(vpnConfig.PacketLogsRetention),
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	case http.MethodPost:
		var (
			writeVPNConfig       bool
			rewriteClientConfigs bool
			setupRequest         VPNSetupRequest
		)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&setupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("setup request decode error: %s", err), http.StatusBadRequest)
			return
		}
		if strings.Join(vpnConfig.ClientRoutes, ", ") != setupRequest.Routes {
			networks := strings.Split(setupRequest.Routes, ",")
			validatedNetworks := []string{}
			for _, network := range networks {
				if strings.TrimSpace(network) == "::/0" {
					validatedNetworks = append(validatedNetworks, "::/0")
				} else {
					_, ipnet, err := net.ParseCIDR(strings.TrimSpace(network))
					if err != nil {
						v.returnError(w, fmt.Errorf("client route %s in wrong format: %s", strings.TrimSpace(network), err), http.StatusBadRequest)
						return
					}
					validatedNetworks = append(validatedNetworks, ipnet.String())
				}
			}
			vpnConfig.ClientRoutes = validatedNetworks
			writeVPNConfig = true
			rewriteClientConfigs = true
		}
		if vpnConfig.Endpoint != setupRequest.VPNEndpoint {
			vpnConfig.Endpoint = setupRequest.VPNEndpoint
			writeVPNConfig = true
			rewriteClientConfigs = true
		}
		addressRangeParsed, err := netip.ParsePrefix(setupRequest.AddressRange)
		if err != nil {
			v.returnError(w, fmt.Errorf("AddressRange in wrong format: %s", err), http.StatusBadRequest)
			return
		}
		if addressRangeParsed.String() != vpnConfig.AddressRange.String() {
			vpnConfig.AddressRange = addressRangeParsed
			writeVPNConfig = true
			rewriteClientConfigs = true
		}
		if setupRequest.ClientAddressPrefix != vpnConfig.ClientAddressPrefix {
			vpnConfig.ClientAddressPrefix = setupRequest.ClientAddressPrefix
			writeVPNConfig = true
			rewriteClientConfigs = true
		}
		port, err := strconv.Atoi(setupRequest.Port)
		if err != nil {
			v.returnError(w, fmt.Errorf("port in wrong format: %s", err), http.StatusBadRequest)
			return
		}
		if port != vpnConfig.Port {
			vpnConfig.Port = port
			writeVPNConfig = true
			rewriteClientConfigs = true
		}

		nameservers := strings.Split(setupRequest.Nameservers, ",")
		for k := range nameservers {
			nameservers[k] = strings.TrimSpace(nameservers[k])
		}
		if !reflect.DeepEqual(nameservers, vpnConfig.Nameservers) {
			vpnConfig.Nameservers = nameservers
			writeVPNConfig = true
			rewriteClientConfigs = true
		}
		if setupRequest.ExternalInterface != vpnConfig.ExternalInterface { // don't rewrite client config
			vpnConfig.ExternalInterface = setupRequest.ExternalInterface
			writeVPNConfig = true
		}
		if setupRequest.DisableNAT != vpnConfig.DisableNAT { // don't rewrite client config
			vpnConfig.DisableNAT = setupRequest.DisableNAT
			writeVPNConfig = true
		}
		if setupRequest.EnablePacketLogs != vpnConfig.EnablePacketLogs {
			vpnConfig.EnablePacketLogs = setupRequest.EnablePacketLogs
			writeVPNConfig = true
		}
		packetLogsRention, err := strconv.Atoi(setupRequest.PacketLogsRetention)
		if err != nil || packetLogsRention < 1 {
			v.returnError(w, fmt.Errorf("incorrect packet log retention. Enter a number of days the logs must be kept (minimum 1)"), http.StatusBadRequest)
			return
		}
		if packetLogsRention != vpnConfig.PacketLogsRetention {
			vpnConfig.PacketLogsRetention = packetLogsRention
			writeVPNConfig = true
		}

		// packetlogtypes
		packetLogTypes := []string{}
		for k, enabled := range vpnConfig.PacketLogsTypes {
			if enabled {
				packetLogTypes = append(packetLogTypes, k)
			}
		}
		sort.Strings(setupRequest.PacketLogsTypes)
		sort.Strings(packetLogTypes)
		if !slices.Equal(setupRequest.PacketLogsTypes, packetLogTypes) {
			vpnConfig.PacketLogsTypes = make(map[string]bool)
			for _, v := range setupRequest.PacketLogsTypes {
				if v == "http+https" || v == "dns" || v == "tcp" {
					vpnConfig.PacketLogsTypes[v] = true
				}
			}
			writeVPNConfig = true
		}

		// write vpn config if config has changed
		if writeVPNConfig {
			err = wireguard.WriteVPNConfig(v.Storage, vpnConfig)
			if err != nil {
				v.returnError(w, fmt.Errorf("could write vpn config: %s", err), http.StatusBadRequest)
				return
			}
			err = wireguard.ReloadVPNServerConfig()
			if err != nil {
				v.returnError(w, fmt.Errorf("unable to reload server config: %s", err), http.StatusBadRequest)
				return
			}
		}
		if rewriteClientConfigs {
			// rewrite client configs
			err = wireguard.UpdateClientsConfig(v.Storage)
			if err != nil {
				v.returnError(w, fmt.Errorf("could not update client vpn configs: %s", err), http.StatusBadRequest)
				return
			}
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	default:
		v.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (v *VPN) templateSetupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		clientTemplate, err := wireguard.GetClientTemplate(v.Storage)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not retrieve client template: %s", err), http.StatusBadRequest)
			return
		}
		serverTemplate, err := wireguard.GetServerTemplate(v.Storage)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not retrieve server template: %s", err), http.StatusBadRequest)
			return
		}
		setupRequest := TemplateSetupRequest{
			ClientTemplate: string(clientTemplate),
			ServerTemplate: string(serverTemplate),
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	case http.MethodPost:
		var templateSetupRequest TemplateSetupRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&templateSetupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("setup request decode error: %s", err), http.StatusBadRequest)
			return
		}
		clientTemplate, err := wireguard.GetClientTemplate(v.Storage)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not retrieve client template: %s", err), http.StatusBadRequest)
			return
		}
		serverTemplate, err := wireguard.GetServerTemplate(v.Storage)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not retrieve server template: %s", err), http.StatusBadRequest)
			return
		}
		if string(clientTemplate) != templateSetupRequest.ClientTemplate {
			err = wireguard.WriteClientTemplate(v.Storage, []byte(templateSetupRequest.ClientTemplate))
			if err != nil {
				v.returnError(w, fmt.Errorf("WriteClientTemplate error: %s", err), http.StatusBadRequest)
				return
			}
			// rewrite client configs
			err = wireguard.UpdateClientsConfig(v.Storage)
			if err != nil {
				v.returnError(w, fmt.Errorf("could not update client vpn configs: %s", err), http.StatusBadRequest)
				return
			}
		}
		if string(serverTemplate) != templateSetupRequest.ServerTemplate {
			err = wireguard.WriteServerTemplate(v.Storage, []byte(templateSetupRequest.ServerTemplate))
			if err != nil {
				v.returnError(w, fmt.Errorf("WriteServerTemplate error: %s", err), http.StatusBadRequest)
				return
			}
		}
		out, err := json.Marshal(templateSetupRequest)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	default:
		v.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (v *VPN) restartVPNHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		v.returnError(w, fmt.Errorf("unsupported method"), http.StatusBadRequest)
		return
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(r.Method, "http://"+wireguard.CONFIGMANAGER_URI+"/restart-vpn", nil)
	if err != nil {
		v.returnError(w, fmt.Errorf("restart request error: %s", err), http.StatusBadRequest)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		v.returnError(w, fmt.Errorf("restart error: %s", err), http.StatusBadRequest)
		return
	}
	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			v.returnError(w, fmt.Errorf("restart error: got status code: %d. Response: %s", resp.StatusCode, bodyBytes), http.StatusBadRequest)
			return
		}
		v.returnError(w, fmt.Errorf("restart error: got status code: %d. Couldn't get response", resp.StatusCode), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close() //nolint:errcheck
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		v.returnError(w, fmt.Errorf("body read error: %s", err), http.StatusBadRequest)
		return
	}

	v.write(w, bodyBytes)
}
