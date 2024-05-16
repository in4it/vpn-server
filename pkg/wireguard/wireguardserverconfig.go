package wireguard

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/in4it/wireguard-server/pkg/storage"
)

func WriteWireGuardServerConfig(storage storage.Iface) error {
	configfileBytes, err := generateWireGuardServerConfig(storage)
	if err != nil {
		return fmt.Errorf("could not generate wireguard server config: %s", err)
	}
	err = os.WriteFile(WIREGUARD_CONFIG, configfileBytes, 0600)
	if err != nil {
		return fmt.Errorf("could not write wireguard config file vpn.conf: %s", err)
	}

	return nil
}

func generateWireGuardServerConfig(storage storage.Iface) ([]byte, error) {
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

	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get vpn config: %s", err)
	}
	privateKey, err := storage.ReadFile(path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %s", err)
	}
	vpnServerData := VPNServerData{
		Address:           vpnConfig.AddressRange.String(),
		PrivateKey:        string(privateKey),
		Port:              vpnConfig.Port,
		DisableNAT:        vpnConfig.DisableNAT,
		ExternalInterface: vpnConfig.ExternalInterface,
	}

	templateContents, err := storage.ReadFile(templatefile)
	if err != nil {
		return nil, fmt.Errorf("cannot read template file (%s): %s", templatefile, err)
	}
	tmpl, err := template.New(path.Base(templatefile)).Parse(string(templateContents))
	if err != nil {
		return nil, fmt.Errorf("could not parse client template: %s", err)
	}

	out := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(out, vpnServerData)
	if err != nil {
		return nil, fmt.Errorf("could not parse client template (execute parsing): %s", err)
	}
	return out.Bytes(), nil
}
