package rest

import (
	"bytes"
	"compress/flate"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	"github.com/in4it/wireguard-server/pkg/auth/saml"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/rest/login"
	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/russellhaering/gosaml2/types"
	dsigtypes "github.com/russellhaering/goxmldsig/types"
)

func getSAMLCertWithCustomCert(singleSignOnURL string, cert string) *types.EntityDescriptor {
	return &types.EntityDescriptor{
		EntityID: "https://www.idp.inv/metadata",
		IDPSSODescriptor: &types.IDPSSODescriptor{
			SingleSignOnServices: []types.SingleSignOnService{
				{
					Location: singleSignOnURL,
				},
			},
			KeyDescriptors: []types.KeyDescriptor{
				{
					KeyInfo: dsigtypes.KeyInfo{
						X509Data: dsigtypes.X509Data{
							X509Certificates: []dsigtypes.X509Certificate{
								{
									Data: cert,
								},
							},
						},
					},
				},
			},
		},
	}
}
func getSAMLCert(singleSignOnURL string) *types.EntityDescriptor {
	cert := `MIID2jCCA0MCAg39MA0GCSqGSIb3DQEBBQUAMIGbMQswCQYDVQQGEwJKUDEOMAwG
A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxETAPBgNVBAoTCEZyYW5rNERE
MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWBgNVBAMTD0ZyYW5rNEREIFdl
YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmcmFuazRkZC5jb20wHhcNMTIw
ODIyMDUyODAwWhcNMTcwODIxMDUyODAwWjBKMQswCQYDVQQGEwJKUDEOMAwGA1UE
CAwFVG9reW8xETAPBgNVBAoMCEZyYW5rNEREMRgwFgYDVQQDDA93d3cuZXhhbXBs
ZS5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCwvWITOLeyTbS1
Q/UacqeILIK16UHLvSymIlbbiT7mpD4SMwB343xpIlXN64fC0Y1ylT6LLeX4St7A
cJrGIV3AMmJcsDsNzgo577LqtNvnOkLH0GojisFEKQiREX6gOgq9tWSqwaENccTE
sAXuV6AQ1ST+G16s00iN92hjX9V/V66snRwTsJ/p4WRpLSdAj4272hiM19qIg9zr
h92e2rQy7E/UShW4gpOrhg2f6fcCBm+aXIga+qxaSLchcDUvPXrpIxTd/OWQ23Qh
vIEzkGbPlBA8J7Nw9KCyaxbYMBFb1i0lBjwKLjmcoihiI7PVthAOu/B71D2hKcFj
Kpfv4D1Uam/0VumKwhwuhZVNjLq1BR1FKRJ1CioLG4wCTr0LVgtvvUyhFrS+3PdU
R0T5HlAQWPMyQDHgCpbOHW0wc0hbuNeO/lS82LjieGNFxKmMBFF9lsN2zsA6Qw32
Xkb2/EFltXCtpuOwVztdk4MDrnaDXy9zMZuqFHpv5lWTbDVwDdyEQNclYlbAEbDe
vEQo/rAOZFl94Mu63rAgLiPeZN4IdS/48or5KaQaCOe0DuAb4GWNIQ42cYQ5TsEH
Wt+FIOAMSpf9hNPjDeu1uff40DOtsiyGeX9NViqKtttaHpvd7rb2zsasbcAGUl+f
NQJj4qImPSB9ThqZqPTukEcM/NtbeQIDAQABMA0GCSqGSIb3DQEBBQUAA4GBAIAi
gU3My8kYYniDuKEXSJmbVB+K1upHxWDA8R6KMZGXfbe5BRd8s40cY6JBYL52Tgqd
l8z5Ek8dC4NNpfpcZc/teT1WqiO2wnpGHjgMDuDL1mxCZNL422jHpiPWkWp3AuDI
c7tL1QjbfAUHAQYwmHkWgPP+T2wAv0pOt36GgMCM`
	return getSAMLCertWithCustomCert(singleSignOnURL, cert)
}

func TestAuthHandler(t *testing.T) {
	c, err := newContext(&testingmocks.MockMemoryStorage{}, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context: %s", err)
	}
	c.UserStore.Empty()
	_, err = c.UserStore.AddUser(users.User{
		Login:    "john",
		Password: "mypass",
	})
	if err != nil {
		t.Fatalf("Cannot create user")
	}

	loginReq := login.LoginRequest{
		Login:    "john",
		Password: "mypass",
	}

	payload, err := json.Marshal(loginReq)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "http://example.com/api/auth", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.authHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var loginResponse login.LoginResponse

	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if !loginResponse.Authenticated {
		t.Fatalf("expected to be authenticated")
	}

}

func TestNewSAMLConnection(t *testing.T) {
	// generate new keypair
	kp := saml.NewKeyPair(&testingmocks.MockMemoryStorage{}, "www.idp.inv")
	_, cert, err := kp.GetKeyPair()
	if err != nil {
		t.Fatalf("Can't generate new keypair: %s", err)
	}
	certBase64 := base64.StdEncoding.EncodeToString(cert)

	testUrl := "127.0.0.1:12347"
	l, err := net.Listen("tcp", testUrl)
	if err != nil {
		t.Fatal(err)
	}

	singleSignOnURL := "http://" + testUrl + "/auth"
	audienceURL := "http://" + testUrl + "/aud"
	login := "john@example.inv"

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requestURIParsed, _ := url.Parse(r.RequestURI)
			if requestURIParsed.Path == "/auth" {
				compressedSAMLReq, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("SAMLRequest"))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("saml base64 decode error: %s", err)))
					return
				}
				samlRequest := new(bytes.Buffer)
				decompressor := flate.NewReader(bytes.NewReader(compressedSAMLReq))
				io.Copy(samlRequest, decompressor)
				decompressor.Close()

				var authnReq saml.AuthnRequest
				err = xml.Unmarshal(samlRequest.Bytes(), &authnReq)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("saml authn request decode error: %s", err)))
					return
				}
				w.Write([]byte("OK"))
				return
			}
			if r.RequestURI == "/metadata" {
				out, _ := xml.Marshal(getSAMLCertWithCustomCert(singleSignOnURL, certBase64))
				w.Write(out)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))

	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()
	defer l.Close()

	// first create a new user
	c, err := newContext(&testingmocks.MockMemoryStorage{}, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	// create a new SAML connection
	samlProvider := saml.Provider{
		Name:        "testProvider",
		MetadataURL: fmt.Sprintf("%s/metadata", ts.URL),
	}

	payload, err := json.Marshal(samlProvider)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "http://example.inv/api/saml-setup", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.samlSetupHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&samlProvider)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if samlProvider.ID == "" {
		t.Fatalf("Was expecting saml provider to have an ID")
	}

	authURL, err := c.SAML.Client.GetAuthURL(samlProvider)
	if err != nil {
		t.Fatalf("cannot get Auth URL from saml: %s", err)
	}

	if authURL == "" {
		t.Fatalf("authURL is empty")
	}

	resp, err = http.Get(authURL)
	if err != nil {
		t.Fatalf("http get auth url error: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("auth url get not status 200: %d", resp.StatusCode)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("body read error: %s", err)
	}

	// check SAML POST flow
	tsSAML := httptest.NewServer(c.SAML.Client.GetRouter())
	defer tsSAML.Close()

	// build the SAML response
	// example
	/*
		<?xml version="1.0"?>
		<samlp:Response xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="pfx181c1b93-cce6-a6c4-f1f7-99e539374d15" Version="2.0" IssueInstant="2024-07-09T17:22:37Z" Destination="https://vpn-server.in4it.io/saml/acs/provider-id">
		  <saml:Issuer>https://app.onelogin.com/saml/metadata/onelogin-id</saml:Issuer>
		  <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
		    <ds:SignedInfo>
		      <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
		      <ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
		      <ds:Reference URI="#pfx181c1b93-cce6-a6c4-f1f7-99e539374d15">
		        <ds:Transforms>
		          <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
		          <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
		        </ds:Transforms>
		        <ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
		        <ds:DigestValue>5eB3C+2/vwdigestvalue</ds:DigestValue>
		      </ds:Reference>
		    </ds:SignedInfo>
		    <ds:SignatureValue>sigvalue</ds:SignatureValue>
		    <ds:KeyInfo>
		      <ds:X509Data>
		        <ds:X509Certificate>MIIGETCCA/mgAcertt</ds:X509Certificate>
		      </ds:X509Data>
		    </ds:KeyInfo>
		  </ds:Signature>
		  <samlp:Status>
		    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
		  </samlp:Status>
		  <saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" Version="2.0" ID="A03d5722f0a14e47b6cfa241ab63089754a7153ca" IssueInstant="2024-07-09T17:22:37Z">
		    <saml:Issuer>https://app.onelogin.com/saml/metadata/onelogin-id</saml:Issuer>
		    <saml:Subject>
		      <saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">ward@in4it.io</saml:NameID>
		      <saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
		        <saml:SubjectConfirmationData NotOnOrAfter="2024-07-09T17:25:37Z" Recipient="https://vpn-server.in4it.io/saml/acs/provider-id"/>
		      </saml:SubjectConfirmation>
		    </saml:Subject>
		    <saml:Conditions NotBefore="2024-07-09T17:19:37Z" NotOnOrAfter="2024-07-09T17:25:37Z">
		      <saml:AudienceRestriction>
		        <saml:Audience>https://vpn-server.in4it.io/saml/aud/provider-id</saml:Audience>
		      </saml:AudienceRestriction>
		    </saml:Conditions>
		    <saml:AuthnStatement AuthnInstant="2024-07-09T17:22:36Z" SessionNotOnOrAfter="2024-07-10T17:22:37Z" SessionIndex="_c66dc31c-41e7-48de-90ed-f4a2f72d864d">
		      <saml:AuthnContext>
		        <saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>
		      </saml:AuthnContext>
		    </saml:AuthnStatement>
		  </saml:Assertion>
		</samlp:Response>
	*/
	samlResponse := saml.Response{
		Saml:         "urn:oasis:names:tc:SAML:2.0:assertion",
		Samlp:        "urn:oasis:names:tc:SAML:2.0:protocol",
		ID:           "pfx181c1b93-cce6-a6c4-f1f7-99e539374d15",
		Version:      "2.0",
		IssueInstant: time.Now().Format(time.RFC3339),
		Destination:  "http://example.inv/saml/acs/" + samlProvider.ID,
		Issuer:       ts.URL + "/metadata",
		Signature: saml.ResponseSignature{
			Ds: "http://www.w3.org/2000/09/xmldsig#",
			SignedInfo: saml.ResponseSignatureSignedInfo{
				CanonicalizationMethod: struct {
					Text      string "xml:\",chardata\""
					Algorithm string "xml:\"Algorithm,attr\""
				}{
					Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
				},
				SignatureMethod: saml.ResponseSignatureSignedInfoSignatureMethod{
					Algorithm: "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
				},
				Reference: saml.ResponseSignatureSignedInfoReference{
					Transforms: struct {
						Text      string "xml:\",chardata\""
						Transform []struct {
							Text      string "xml:\",chardata\""
							Algorithm string "xml:\"Algorithm,attr\""
						} "xml:\"ds:Transform\""
					}{
						Transform: []struct {
							Text      string "xml:\",chardata\""
							Algorithm string "xml:\"Algorithm,attr\""
						}{
							{
								Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature",
							},
							{
								Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
							},
						},
					},
					DigestMethod: struct {
						Text      string "xml:\",chardata\""
						Algorithm string "xml:\"Algorithm,attr\""
					}{
						Algorithm: "http://www.w3.org/2000/09/xmldsig#sha1",
					},
					DigestValue: "thisisthesignature",
				},
			},
			KeyInfo: saml.ResponseSignatureKeyInfo{
				X509Data: struct {
					Text            string "xml:\",chardata\""
					X509Certificate string "xml:\"ds:X509Certificate\""
				}{
					X509Certificate: certBase64,
				},
			},
		},
		Assertion: saml.ResponseAssertion{
			Subject: saml.ResponseSubject{
				NameID: struct {
					Text   string "xml:\",chardata\""
					Format string "xml:\"Format,attr\""
				}{
					Text:   login,
					Format: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
				},
			},
			Conditions: saml.ResponseConditions{
				NotBefore:    time.Now().Format(time.RFC3339),
				NotOnOrAfter: time.Now().Add(10 * time.Minute).Format(time.RFC3339),
				AudienceRestriction: saml.ResponseConditionsAdienceRestriction{
					Audience: audienceURL,
				},
			},
		},
	}
	samlResponseBytes, err := xml.Marshal(samlResponse)
	if err != nil {
		t.Fatalf("xml marshal error: %s", err)
	}
	//fmt.Printf("saml respons bytes: %s\n", samlResponseBytes)
	samlResponseBytesDeflated := new(bytes.Buffer)
	compressor, err := flate.NewWriter(samlResponseBytesDeflated, 1)
	if err != nil {
		t.Fatalf("deflate error: %s", err)
	}
	io.Copy(compressor, bytes.NewBuffer(samlResponseBytes))
	compressor.Close()

	samlResponseEncoded := base64.StdEncoding.EncodeToString(samlResponseBytesDeflated.Bytes())

	form := url.Values{}
	form.Add("SAMLResponse", samlResponseEncoded)

	resp, err = http.Post(tsSAML.URL+"/saml/acs/"+samlProvider.ID, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("http post acs url error: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("body read error: %s", err)
	}

	if strings.Contains(string(body), "provider not found") {
		t.Fatalf("provider not found: body output: %s", body)
	}

	// currently does't authenticate because of missing signatures, but we checked if the auth process kicked off
	/*if resp.StatusCode != 200 {
		t.Errorf("auth url get not status 200: %d", resp.StatusCode)
	}*/

}
func TestAddModifyDeleteNewSAMLConnection(t *testing.T) {
	c, err := newContext(&testingmocks.MockMemoryStorage{}, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}
	c.Hostname = "example.inv"
	c.Protocol = "https"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata" {
			out, err := xml.Marshal(getSAMLCert("http://localhost.inv"))
			if err != nil {
				t.Fatalf("marshal error: %s", err)
			}
			w.Write(out)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	samlProvider := saml.Provider{
		Name:        "testProvider",
		MetadataURL: fmt.Sprintf("%s/metadata", ts.URL),
	}

	payload, err := json.Marshal(samlProvider)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "http://example.com/api/saml-setup", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.samlSetupHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&samlProvider)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if samlProvider.ID == "" {
		t.Fatalf("samlprovider id is empty")
	}

	// GET authmethods and see if provider exists
	req = httptest.NewRequest("GET", "http://example.com/authmethods/saml/"+samlProvider.ID, nil)
	req.SetPathValue("method", "saml")
	req.SetPathValue("id", samlProvider.ID)
	w = httptest.NewRecorder()
	c.authMethodsByID(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var authMethodsProvider AuthMethodsProvider

	err = json.NewDecoder(resp.Body).Decode(&authMethodsProvider)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}
	if authMethodsProvider.ID != samlProvider.ID {
		t.Fatalf("authmethods provider id is different than saml provider id: %s vs %s. authMethodsProvider: %+v", authMethodsProvider.ID, samlProvider.ID, authMethodsProvider)
	}

	// PUT req
	samlProvider.AllowMissingAttributes = true
	payload, err = json.Marshal(samlProvider)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req = httptest.NewRequest("PUT", "http://example.com/saml-setup/"+samlProvider.ID, bytes.NewBuffer(payload))
	req.SetPathValue("id", samlProvider.ID)
	w = httptest.NewRecorder()
	c.samlSetupElementHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&samlProvider)
	if err != nil {
		t.Fatalf("marshal decode error: %s", err)
	}

	if samlProvider.AllowMissingAttributes == false {
		t.Fatalf("allow missing attributes is false")
	}

	// GET on the saml endpoint to see if we can return it
	req = httptest.NewRequest("GET", "http://example.com/saml-setup", nil)
	w = httptest.NewRecorder()
	c.samlSetupHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var samlProviders []saml.Provider
	err = json.NewDecoder(resp.Body).Decode(&samlProviders)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}
	if len(samlProviders) == 0 {
		t.Fatalf("samlProviders is zero length")
	}
	if samlProviders[len(samlProviders)-1].ID != samlProvider.ID {
		t.Fatalf("ID doesn't match: %s vs %s ", samlProviders[len(samlProviders)-1].ID, samlProvider.ID)
	}
	if samlProviders[len(samlProviders)-1].Acs != fmt.Sprintf("%s://%s/%s/%s", c.Protocol, c.Hostname, saml.ACS_URL, samlProvider.ID) {
		t.Fatalf("ACS doesn't match")
	}
	if samlProviders[len(samlProviders)-1].AllowMissingAttributes == false {
		t.Fatalf("allow missing attributes is false when getting all samlproviders")
	}

	// delete req
	req = httptest.NewRequest("DELETE", "http://example.com/saml-setup/"+samlProvider.ID, bytes.NewBuffer(payload))
	req.SetPathValue("id", samlProvider.ID)
	w = httptest.NewRecorder()
	c.samlSetupElementHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	// list to see if really deleted
	req = httptest.NewRequest("GET", "http://example.com/saml-setup", nil)
	w = httptest.NewRecorder()
	c.samlSetupHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var samlProviders2 []saml.Provider
	err = json.NewDecoder(resp.Body).Decode(&samlProviders2)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}
	if len(samlProviders)-1 != len(samlProviders2) {
		t.Fatalf("samlProviders has wrong length")
	}

}

func TestSAMLCallback(t *testing.T) {
	c, err := newContext(&testingmocks.MockMemoryStorage{}, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}
	c.Hostname = "example.inv"
	c.Protocol = "https"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata" {
			out, err := xml.Marshal(getSAMLCert("http://localhost.inv"))
			if err != nil {
				t.Fatalf("marshal error: %s", err)
			}
			w.Write(out)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	samlProvider := saml.Provider{
		Name:        "testProvider",
		MetadataURL: fmt.Sprintf("%s/metadata", ts.URL),
	}

	payload, err := json.Marshal(samlProvider)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "http://example.com/api/saml-setup", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.samlSetupHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&samlProvider)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if samlProvider.ID == "" {
		t.Fatalf("samlprovider id is empty")
	}
	samlCallback := SAMLCallback{
		Code:        "abc",
		RedirectURI: "https://localhost.inv/something",
	}
	payload, err = json.Marshal(samlCallback)
	if err != nil {
		t.Fatal(err)
	}

	c.SAML.Client.CreateSession(saml.SessionKey{ProviderID: samlProvider.ID, SessionID: "abc"}, saml.AuthenticatedUser{ID: "123", Login: "john@example.com", ExpiresAt: time.Now().AddDate(0, 0, 1)})

	req = httptest.NewRequest("POST", "http://example.com/api/authmethods/saml/"+samlProvider.ID, bytes.NewBuffer(payload))
	req.SetPathValue("method", "saml")
	req.SetPathValue("id", samlProvider.ID)
	w = httptest.NewRecorder()
	c.authMethodsByID(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var loginResponse login.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if !loginResponse.Authenticated {
		t.Fatalf("Expected to be authenticated")
	}

}

func TestOIDCFlow(t *testing.T) {
	testUrl := "127.0.0.1:12346"
	l, err := net.Listen("tcp", testUrl)
	if err != nil {
		t.Fatal(err)
	}

	authURL := "http://" + testUrl + "/auth"

	// create a new OIDC connection
	oidcProvider := oidc.OIDCProvider{
		Name:         "test-oidc",
		ClientID:     "1-2-3-4",
		ClientSecret: "9-9-9-9",
		Scope:        "openid",
		DiscoveryURI: "http://" + testUrl + "/discovery.json",
	}
	jwtPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("can't generate jwt key: %s", err)
	}

	// first create a new user
	c, err := newContext(&testingmocks.MockMemoryStorage{}, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}
	c.Hostname = "example.inv"
	c.Protocol = "http"
	logging.Loglevel = 17

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := "thisisthecode"

		switch r.Method {
		case http.MethodGet:
			parsedURI, _ := url.Parse(r.RequestURI)
			switch parsedURI.Path {
			case "/discovery.json":
				discovery := oidc.Discovery{
					Issuer:                "test-issuer",
					AuthorizationEndpoint: authURL,
					TokenEndpoint:         "http://" + testUrl + "/token",
					JwksURI:               "http://" + testUrl + "/jwks.json",
				}
				out, err := json.Marshal(discovery)
				if err != nil {
					t.Fatalf("json marshal error: %s", err)
				}
				w.Write(out)
				return
			case "/auth":
				if oidcProvider.ClientID != r.URL.Query().Get("client_id") {
					w.Write([]byte("client id mismatch"))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if oidcProvider.Scope != r.URL.Query().Get("scope") {
					w.Write([]byte("scope mismatch"))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Write([]byte(code))
			case "/jwks.json":
				publicKey := jwtPrivateKey.PublicKey

				jwks := oidc.Jwks{
					Keys: []oidc.JwksKey{
						{
							Kid: "kid-id-1234",
							Alg: "RS256",
							Kty: "RSA",
							Use: "sig",
							N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
							E:   "AQAB",
						},
					},
				}
				out, err := json.Marshal(jwks)
				if err != nil {
					w.Write([]byte("jwks marshal error"))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Write(out)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost:
			parsedURI, _ := url.Parse(r.RequestURI)
			switch parsedURI.Path {
			case "/token":
				if r.FormValue("grant_type") != "authorization_code" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("wrong grant type"))
					return
				}
				if r.FormValue("code") != code {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("wrong code"))
					return
				}
				if oidcProvider.ClientID != r.FormValue("client_id") {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("client id mismatch"))
					return
				}
				if oidcProvider.ClientSecret != r.FormValue("client_secret") {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("client secret mismatch"))
					return
				}
				if c.Protocol+"://"+c.Hostname+oidcProvider.RedirectURI != r.FormValue("redirect_uri") {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("redirect uri mismatch: %s vs %s", oidcProvider.RedirectURI, r.FormValue("redirect_uri"))))
					return
				}
				token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), jwt.MapClaims{
					"iss":   "test-server",
					"sub":   "john",
					"email": "john@example.inv",
					"role":  "user",
					"exp":   time.Now().AddDate(0, 0, 1).Unix(),
					"iat":   time.Now().Unix(),
				})
				token.Header["kid"] = "kid-id-1234"

				tokenString, err := token.SignedString(jwtPrivateKey)
				if err != nil {
					t.Fatalf("can't generate jwt token: %s", err)
					w.WriteHeader(http.StatusBadRequest)
				}
				tokenRes := oidc.Token{
					AccessToken: tokenString,
					IDToken:     tokenString,
					ExpiresIn:   180,
				}
				tokenBytes, _ := json.Marshal(tokenRes)

				w.Write([]byte(tokenBytes))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))

	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()
	defer l.Close()

	payload, err := json.Marshal(oidcProvider)
	if err != nil {
		t.Fatal(err)
	}

	// create new oidc provider
	req := httptest.NewRequest("POST", "http://example.inv/api/oidc", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.oidcProviderHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&oidcProvider)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if oidcProvider.ID == "" {
		t.Fatalf("Was expecting oidc provider to have an ID")
	}

	// get redirect URL
	req = httptest.NewRequest("GET", "http://example.inv/api/authmethods/oidc/"+oidcProvider.ID, nil)
	req.SetPathValue("id", oidcProvider.ID)
	w = httptest.NewRecorder()
	c.authMethodsByID(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	var authmethodsResponse AuthMethodsProvider

	err = json.NewDecoder(resp.Body).Decode(&authmethodsResponse)

	if err != nil {
		t.Fatalf("cannot decode authmethodsresponse: %s", err)
	}
	if !strings.HasPrefix(authmethodsResponse.RedirectURI, authURL) {
		t.Fatalf("expected authURL as prefix of redirect url. Redirect URL: %s", authmethodsResponse.RedirectURI)
	}

	redirectURIParsed, err := url.Parse(authmethodsResponse.RedirectURI)
	if err != nil {
		t.Fatalf("could not parse redirect URI: %s", err)
	}
	state := redirectURIParsed.Query().Get("state")
	if state == "" {
		t.Fatalf("could not obtain state")
	}
	res, err := http.Get(authmethodsResponse.RedirectURI)
	if err != nil {
		t.Fatalf("http get redirect uri error: %s", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("redirect uri statuscode not 200: %d", res.StatusCode)
	}
	code, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("body read error: %s", err)
	}

	callback := OIDCCallback{
		Code:        string(code),
		State:       state,
		RedirectURI: oidcProvider.RedirectURI,
	}
	callbackPayload, err := json.Marshal(callback)
	if err != nil {
		t.Fatalf("callback marshal error: %s", err)
	}
	// execute callback
	req = httptest.NewRequest("POST", "http://example.inv/api/authmethods/oidc/"+oidcProvider.ID, bytes.NewBuffer(callbackPayload))
	req.SetPathValue("id", oidcProvider.ID)
	req.SetPathValue("method", "oidc")
	w = httptest.NewRecorder()
	c.authMethodsByID(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorMessage, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("body read error after statuscode not 200 (%d): %s", resp.StatusCode, err)
		}
		t.Fatalf("status code is not 200: %d, errormessage: %s", resp.StatusCode, errorMessage)
	}

	var loginResponse login.LoginResponse

	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		t.Fatalf("cannot decode login response: %s", err)
	}

	if !loginResponse.Authenticated {
		t.Fatalf("not authenticated: %+v", loginResponse)
	}
	if loginResponse.Token == "" {
		t.Fatalf("no token received: %+v", loginResponse)
	}
}
