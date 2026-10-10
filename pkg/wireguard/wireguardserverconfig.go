package wireguard

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard/network"
)

// variable so tests can stub it
var getInterfaceDefaultGw = network.GetInterfaceDefaultGw

// interface names end up in the PostUp/PostDown shell commands of the server
// config (executed as root by wg-quick), so only allow safe characters.
// Linux interface names are at most 15 characters.
var interfaceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,15}$`)

// ValidateExternalInterface checks the external interface that is used for
// NAT. An empty interface is only allowed when NAT is disabled.
func ValidateExternalInterface(externalInterface string, disableNAT bool) error {
	if externalInterface == "" {
		if disableNAT {
			return nil
		}
		return fmt.Errorf("external interface can't be empty when NAT is enabled")
	}
	if !interfaceNameRegex.MatchString(externalInterface) {
		return fmt.Errorf("invalid external interface name: %q", externalInterface)
	}
	return nil
}

// getExternalInterface returns the external interface to render in the server
// config. When it's not set (e.g. the network wasn't up yet when the config was
// created at first boot), it's detected again, so the server config never
// contains "-o  -j MASQUERADE".
func getExternalInterface(vpnConfig VPNConfig) (string, error) {
	externalInterface := strings.TrimSpace(vpnConfig.ExternalInterface)
	if externalInterface == "" && !vpnConfig.DisableNAT {
		detected, err := getInterfaceDefaultGw()
		if err != nil {
			return "", fmt.Errorf("external interface is not set and could not be detected (%s). Set the external interface in the VPN settings or disable NAT", err)
		}
		externalInterface = detected
	}
	if err := ValidateExternalInterface(externalInterface, vpnConfig.DisableNAT); err != nil {
		return "", err
	}
	return externalInterface, nil
}

func WriteWireGuardServerConfig(storage storage.Iface) error {
	configfileBytes, err := generateWireGuardServerConfig(storage)
	if err != nil {
		return fmt.Errorf("could not generate wireguard server config: %w", err)
	}
	err = os.WriteFile(WIREGUARD_CONFIG, configfileBytes, 0600)
	if err != nil {
		return fmt.Errorf("could not write wireguard config file vpn.conf: %s", err)
	}

	return nil
}

func GetServerTemplate(storage storage.Iface) ([]byte, error) {
	templatefile := storage.ConfigPath(path.Join(WIREGUARD_TEMPLATE_DIR, WIREGUARD_TEMPLATE_SERVER))
	err := storage.EnsurePath(storage.ConfigPath(WIREGUARD_TEMPLATE_DIR))
	if err != nil {
		return nil, fmt.Errorf("cannot ensure path for template directory: %s", err)
	}
	err = storage.EnsureOwnership(storage.ConfigPath(WIREGUARD_TEMPLATE_DIR), "vpn")
	if err != nil {
		return nil, fmt.Errorf("cannot ensure path for template directory: %s", err)
	}

	if !storage.FileExists(templatefile) {
		err := storage.WriteFile(templatefile, []byte(DEFAULT_SERVER_TEMPLATE))
		if err != nil {
			return nil, fmt.Errorf("could not create template (%s): %s", templatefile, err)
		}
		err = storage.EnsureOwnership(templatefile, "vpn")
		if err != nil {
			return nil, fmt.Errorf("cannot ensure ownership for template file (%s): %s", templatefile, err)
		}
	}

	templateContents, err := storage.ReadFile(templatefile)
	if err != nil {
		return nil, fmt.Errorf("cannot read template file (%s): %s", templatefile, err)
	}
	return templateContents, nil
}

func WriteServerTemplate(storage storage.Iface, body []byte) error {
	templatefile := storage.ConfigPath(path.Join(WIREGUARD_TEMPLATE_DIR, WIREGUARD_TEMPLATE_SERVER))

	err := storage.WriteFile(templatefile, body)
	if err != nil {
		return fmt.Errorf("could not write template (%s): %s", templatefile, err)
	}

	return nil
}

// CheckWireGuardServerConfig renders the server config without writing it, to
// find out whether the vpn can be started with the current config and template.
func CheckWireGuardServerConfig(storage storage.Iface) error {
	_, err := generateWireGuardServerConfig(storage)
	return err
}

func generateWireGuardServerConfig(storage storage.Iface) ([]byte, error) {
	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get vpn config: %s", err)
	}
	privateKey, err := storage.ReadFile(path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %s", err)
	}
	externalInterface, err := getExternalInterface(vpnConfig)
	if err != nil {
		return nil, err
	}
	vpnServerData := VPNServerData{
		Address:           vpnConfig.AddressRange.String(),
		PrivateKey:        string(privateKey),
		Port:              vpnConfig.Port,
		DisableNAT:        vpnConfig.DisableNAT,
		ExternalInterface: externalInterface,
	}

	templateContents, err := GetServerTemplate(storage)
	if err != nil {
		return nil, fmt.Errorf("cannot get template file: %s", err)
	}
	tmpl, err := parseServerTemplate(templateContents)
	if err != nil {
		return nil, fmt.Errorf("could not parse client template: %s", err)
	}

	out := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(out, vpnServerData)
	if err != nil {
		return nil, fmt.Errorf("could not parse client template (execute parsing): %s", err)
	}
	// check the rendered config, not only the template (e.g. the template can
	// render differently depending on the data)
	if err := checkServerConfig(out.Bytes(), externalInterface); err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}
	return out.Bytes(), nil
}
