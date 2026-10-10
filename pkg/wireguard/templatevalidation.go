package wireguard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"
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
// The rendered config can only contain the PostUp/PostDown commands that the
// default template generates (see checkServerConfig).
func ValidateServerTemplate(contents []byte) error {
	tmpl, err := parseServerTemplate(contents)
	if err != nil {
		return fmt.Errorf("could not parse server template: %s", err)
	}
	for _, disableNAT := range []bool{false, true} {
		data := VPNServerData{
			Address:           "10.189.184.1/21",
			PrivateKey:        "private-key",
			Port:              51820,
			Clients:           []VPNServerClient{{PublicKey: "public-key", AllowedIPs: "10.189.184.2/32", PresharedKey: "preshared-key"}},
			DisableNAT:        disableNAT,
			ExternalInterface: "eth0",
		}
		out := bytes.NewBuffer([]byte{})
		err = tmpl.Execute(out, data)
		if err != nil {
			return fmt.Errorf("could not render server template: %s", err)
		}
		if err := checkServerConfig(out.Bytes(), data.ExternalInterface); err != nil {
			return err
		}
	}
	return nil
}

// wg-quick keys that run commands, or change the routing, dns or the
// persistence of the config. These keys are only allowed with the values that
// the default server template generates.
var restrictedServerConfigKeys = map[string]string{
	"preup":      "PreUp",
	"postup":     "PostUp",
	"predown":    "PreDown",
	"postdown":   "PostDown",
	"saveconfig": "SaveConfig",
	"table":      "Table",
	"dns":        "DNS",
}

func isHookKey(key string) bool {
	return key == "preup" || key == "postup" || key == "predown" || key == "postdown"
}

// parseServerConfigLine splits a line of a wireguard config in a key and a
// value, the same way wg-quick does: comments (#) are removed, the key is
// everything before the first =, the value everything after it (or the whole
// line when there's no =). The key is returned in lowercase, without any
// whitespace, control or format characters. ok is false when the key
// contains non-ascii characters.
func parseServerConfigLine(line string) (key string, value string, ok bool) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}
	rawKey, rawValue := line, line
	if i := strings.IndexByte(line, '='); i >= 0 {
		rawKey, rawValue = line[:i], line[i+1:]
	}
	var normalizedKey strings.Builder
	for _, r := range rawKey {
		if unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		if r > unicode.MaxASCII {
			return "", "", false
		}
		normalizedKey.WriteRune(unicode.ToLower(r))
	}
	return normalizedKey.String(), strings.TrimFunc(rawValue, unicode.IsSpace), true
}

type serverConfigEntry struct {
	key   string
	value string
}

// allowedServerConfigEntries returns the restricted keys and values that the
// default server template generates for the external interface
func allowedServerConfigEntries(externalInterface string) (map[serverConfigEntry]bool, error) {
	tmpl, err := parseServerTemplate([]byte(DEFAULT_SERVER_TEMPLATE))
	if err != nil {
		return nil, fmt.Errorf("could not parse default server template: %s", err)
	}
	out := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(out, VPNServerData{ExternalInterface: externalInterface})
	if err != nil {
		return nil, fmt.Errorf("could not render default server template: %s", err)
	}
	allowed := map[serverConfigEntry]bool{}
	for _, line := range splitConfigLines(out.String()) {
		key, value, ok := parseServerConfigLine(line)
		if !ok {
			continue
		}
		if _, restricted := restrictedServerConfigKeys[key]; restricted {
			allowed[serverConfigEntry{key: key, value: value}] = true
		}
	}
	return allowed, nil
}

func splitConfigLines(config string) []string {
	return strings.FieldsFunc(config, func(r rune) bool { return r == '\n' || r == '\r' })
}

// SERVER_TEMPLATE_DOCS_URL explains which settings the server template supports
const SERVER_TEMPLATE_DOCS_URL = "https://vpn-documentation.in4it.com/faq/#the-server-template-contains-unsupported-settings"

// ErrServerConfigNotAllowed is returned when the rendered server config contains
// settings that aren't supported in the server template
var ErrServerConfigNotAllowed = errors.New("server config not allowed")

// checkServerConfig checks the rendered server config. wg-quick runs the
// PreUp/PostUp/PreDown/PostDown commands of the config, so only the commands
// that the default server template generates (the NAT rules) are allowed.
// SaveConfig, Table and DNS are also only allowed when the default template
// uses them. Keys are matched case-insensitively, in every section.
func checkServerConfig(config []byte, externalInterface string) error {
	allowed, err := allowedServerConfigEntries(externalInterface)
	if err != nil {
		return err
	}
	for i, line := range strings.Split(string(config), "\n") {
		for _, subLine := range splitConfigLines(line) {
			key, value, ok := parseServerConfigLine(subLine)
			if !ok {
				return fmt.Errorf("%w: invalid key in server template (line %d)", ErrServerConfigNotAllowed, i+1)
			}
			name, restricted := restrictedServerConfigKeys[key]
			if !restricted || allowed[serverConfigEntry{key: key, value: value}] {
				continue
			}
			if isHookKey(key) {
				return fmt.Errorf("%w: custom PreUp/PostUp/PreDown/PostDown commands are not supported in the server template (line %d: %s). See %s", ErrServerConfigNotAllowed, i+1, name, SERVER_TEMPLATE_DOCS_URL)
			}
			return fmt.Errorf("%w: %s is not supported in the server template (line %d). See %s", ErrServerConfigNotAllowed, name, i+1, SERVER_TEMPLATE_DOCS_URL)
		}
	}
	return nil
}
