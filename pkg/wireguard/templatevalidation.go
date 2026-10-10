package wireguard

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

func parseClientTemplate(contents []byte) (*template.Template, error) {
	return template.New("client.tmpl").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(string(contents))
}

func parseServerTemplate(contents []byte) (*template.Template, error) {
	return template.New(WIREGUARD_TEMPLATE_SERVER).Parse(string(contents))
}

// ValidateClientTemplate parses and renders the client template with sample
// data, so a broken template is rejected before it's saved.
func ValidateClientTemplate(contents []byte) error {
	tmpl, err := parseClientTemplate(contents)
	if err != nil {
		return fmt.Errorf("could not parse client template: %s", err)
	}
	err = tmpl.Execute(io.Discard, VPNClientData{
		Address:         "10.189.184.2/32",
		DNS:             "1.1.1.1",
		PrivateKey:      "private-key",
		ServerPublicKey: "server-public-key",
		PresharedKey:    "preshared-key",
		Endpoint:        "vpn.example.com:51820",
		AllowedIPs:      []string{"0.0.0.0/0"},
	})
	if err != nil {
		return fmt.Errorf("could not render client template: %s", err)
	}
	return nil
}

// ValidateServerTemplate parses and renders the server template with sample
// data. A server template that doesn't render stops the VPN from starting.
func ValidateServerTemplate(contents []byte) error {
	tmpl, err := parseServerTemplate(contents)
	if err != nil {
		return fmt.Errorf("could not parse server template: %s", err)
	}
	err = tmpl.Execute(io.Discard, VPNServerData{
		Address:           "10.189.184.1/21",
		PrivateKey:        "private-key",
		Port:              51820,
		Clients:           []VPNServerClient{{PublicKey: "public-key", AllowedIPs: "10.189.184.2/32", PresharedKey: "preshared-key"}},
		ExternalInterface: "eth0",
	})
	if err != nil {
		return fmt.Errorf("could not render server template: %s", err)
	}
	return nil
}
