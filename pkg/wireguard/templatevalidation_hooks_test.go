package wireguard

import (
	"path"
	"strings"
	"testing"

	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
)

const testNATPostUp = "PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE"
const testNATPostDown = "PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE"

func TestValidateServerTemplateAllowedSettings(t *testing.T) {
	allowed := map[string]string{
		"default": DEFAULT_SERVER_TEMPLATE,
		"changed address, port, mtu and comments": `# my server template
[Interface]
Address = 10.0.0.1/24 # comment
PrivateKey = {{ .PrivateKey }}
ListenPort = 51821
MTU = 1380
# PostUp = this is a comment
{{if not .DisableNAT }}
` + testNATPostUp + `
` + testNATPostDown + `
{{end}}
`,
		"nat lines without condition, different spacing and case": `[Interface]
Address = {{ .Address }}
PrivateKey = {{ .PrivateKey }}
ListenPort = {{ .Port }}
postup=iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE
  POSTDOWN   =   iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE   
`,
		"no nat": `[Interface]
Address = {{ .Address }}
PrivateKey = {{ .PrivateKey }}
ListenPort = {{ .Port }}
`,
		"peers": DEFAULT_SERVER_TEMPLATE + `
{{ range .Clients }}
[Peer]
PublicKey = {{ .PublicKey }}
PresharedKey = {{ .PresharedKey }}
AllowedIPs = {{ .AllowedIPs }}
{{ end }}
`,
	}
	for name, tmpl := range allowed {
		if err := ValidateServerTemplate([]byte(tmpl)); err != nil {
			t.Errorf("template %q should be valid: %s", name, err)
		}
	}
}

func TestValidateServerTemplateCustomHooks(t *testing.T) {
	base := "[Interface]\nAddress = {{ .Address }}\nPrivateKey = {{ .PrivateKey }}\nListenPort = {{ .Port }}\n"
	rejected := map[string]string{
		"custom PostUp":                base + "PostUp = touch /tmp/test\n",
		"custom PreUp":                 base + "PreUp = touch /tmp/test\n",
		"custom PreDown":               base + "PreDown = touch /tmp/test\n",
		"custom PostDown":              base + "PostDown = touch /tmp/test\n",
		"lowercase":                    base + "postup = touch /tmp/test\n",
		"uppercase":                    base + "POSTUP = touch /tmp/test\n",
		"mixed case and whitespace":    base + " \t pOsTuP\t=touch /tmp/test\n",
		"no equals sign":               base + "PostUp\n",
		"carriage return":              base + "PostUp\r = touch /tmp/test\r\n",
		"null byte in key":             base + "Post\x00Up = touch /tmp/test\n",
		"zero width space in key":      base + "Post\u200bUp = touch /tmp/test\n",
		"non-ascii key":                base + "P\u03bfstup = touch /tmp/test\n", // greek omicron
		"extra command after nat rule": base + strings.Replace(testNATPostUp, "MASQUERADE", "MASQUERADE; touch /tmp/test", 1) + "\n",
		"nat rule as PreUp":            base + strings.Replace(testNATPostUp, "PostUp", "PreUp", 1) + "\n",
		"nat up rule as PostDown":      base + strings.Replace(testNATPostUp, "PostUp", "PostDown", 1) + "\n",
		"hook in second interface":     DEFAULT_SERVER_TEMPLATE + "\n[Interface]\nPostUp = touch /tmp/test\n",
		"hook in peer section":         DEFAULT_SERVER_TEMPLATE + "\n[Peer]\nPostUp = touch /tmp/test\n",
		"hook when nat is disabled":    base + "{{ if .DisableNAT }}PostUp = touch /tmp/test{{ end }}\n",
		"hook from template data":      base + "PostUp = {{ .Address }}\n",
		"SaveConfig":                   base + "SaveConfig = true\n",
		"Table":                        base + "Table = off\n",
		"DNS":                          base + "DNS = 1.1.1.1\n",
		"table lowercase":              base + "table=1234\n",
	}
	for name, tmpl := range rejected {
		if err := ValidateServerTemplate([]byte(tmpl)); err == nil {
			t.Errorf("template %q: expected validation error", name)
		}
	}
	err := ValidateServerTemplate([]byte(base + "PostUp = touch /tmp/test\n"))
	if err == nil || !strings.Contains(err.Error(), "custom PreUp/PostUp/PreDown/PostDown commands are not supported in the server template") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGenerateServerConfigRejectsCustomHooks(t *testing.T) {
	stubConfigManager(t)
	storage := &memorystorage.MockMemoryStorage{}
	vpnConfig, err := CreateNewVPNConfig(storage)
	if err != nil {
		t.Fatalf("CreateNewVPNConfig error: %s", err)
	}
	vpnConfig.ExternalInterface = "ens5"
	if err := WriteVPNConfig(storage, vpnConfig); err != nil {
		t.Fatalf("WriteVPNConfig error: %s", err)
	}
	// default template renders
	out, err := generateWireGuardServerConfig(storage)
	if err != nil {
		t.Fatalf("generate error: %s", err)
	}
	if !strings.Contains(string(out), "PostUp = iptables") || !strings.Contains(string(out), "-o ens5 -j MASQUERADE") {
		t.Fatalf("expected nat rules in config:\n%s", out)
	}
	// a template that only renders a custom command with the real data
	// (passes the validation with sample data) is refused when rendering
	tmpl := DEFAULT_SERVER_TEMPLATE + "{{ if eq .ExternalInterface \"ens5\" }}PostUp = touch /tmp/test{{ end }}\n"
	if err := ValidateServerTemplate([]byte(tmpl)); err != nil {
		t.Fatalf("template should pass validation with sample data: %s", err)
	}
	if err := WriteServerTemplate(storage, []byte(tmpl)); err != nil {
		t.Fatalf("WriteServerTemplate error: %s", err)
	}
	if out, err := generateWireGuardServerConfig(storage); err == nil {
		t.Fatalf("expected error, got config:\n%s", out)
	}
	if err := CheckWireGuardServerConfig(storage); err == nil {
		t.Fatalf("expected CheckWireGuardServerConfig error")
	}
	// a template written directly to the storage (not validated) is refused too
	if err := WriteServerTemplate(storage, []byte(DEFAULT_SERVER_TEMPLATE+"[Interface]\nPreUp = touch /tmp/test\n")); err != nil {
		t.Fatalf("WriteServerTemplate error: %s", err)
	}
	if out, err := generateWireGuardServerConfig(storage); err == nil {
		t.Fatalf("expected error, got config:\n%s", out)
	}
	// data from the storage that ends up in the config is checked as well
	if err := WriteServerTemplate(storage, []byte(DEFAULT_SERVER_TEMPLATE)); err != nil {
		t.Fatalf("WriteServerTemplate error: %s", err)
	}
	if err := storage.WriteFile(path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME), []byte("key\nPostUp = touch /tmp/test")); err != nil {
		t.Fatal(err)
	}
	if out, err := generateWireGuardServerConfig(storage); err == nil {
		t.Fatalf("expected error, got config:\n%s", out)
	}
}
