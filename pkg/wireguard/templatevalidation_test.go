package wireguard

import (
	"testing"
)

func TestValidateDefaultTemplates(t *testing.T) {
	if err := ValidateClientTemplate([]byte(DEFAULT_CLIENT_TEMPLATE)); err != nil {
		t.Fatalf("default client template should be valid: %s", err)
	}
	if err := ValidateServerTemplate([]byte(DEFAULT_SERVER_TEMPLATE)); err != nil {
		t.Fatalf("default server template should be valid: %s", err)
	}
}

func TestValidateInvalidTemplates(t *testing.T) {
	invalid := map[string]string{
		"unknown field":  "[Interface]\nAddress = {{ .Adress }}\n",
		"unclosed":       "[Interface]\nAddress = {{ .Address\n",
		"unknown func":   "{{ Foo .Address }}",
		"bad range type": "{{ range .Address }}{{ end }}",
	}
	for name, tmpl := range invalid {
		if err := ValidateClientTemplate([]byte(tmpl)); err == nil {
			t.Errorf("client template %q: expected error", name)
		}
		if err := ValidateServerTemplate([]byte(tmpl)); err == nil {
			t.Errorf("server template %q: expected error", name)
		}
	}
	// StringsJoin is available in the client template only
	if err := ValidateClientTemplate([]byte(`{{ StringsJoin .AllowedIPs ", " }}`)); err != nil {
		t.Errorf("client template with StringsJoin should be valid: %s", err)
	}
	if err := ValidateServerTemplate([]byte(`{{ range .Clients }}{{ .PublicKey }}{{ end }}`)); err != nil {
		t.Errorf("server template ranging over clients should be valid: %s", err)
	}
}
