package saml

import (
	"encoding/xml"
	"net/http"
	"sync"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
	saml2 "github.com/russellhaering/gosaml2"
)

type Provider struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	MetadataURL            string `json:"metadataURL"`
	Issuer                 string `json:"issuer,omitempty"`
	Audience               string `json:"audience,omitempty"`
	Acs                    string `json:"acs,omitempty"`
	AllowMissingAttributes bool   `json:"allowMissingAttributes,omitempty"`
}

type saml struct {
	Providers       *[]Provider
	serviceProvider map[string]*saml2.SAMLServiceProvider
	sessions        map[SessionKey]AuthenticatedUser
	hostname        *string
	protocol        *string
	mu              sync.Mutex
	storage         storage.Iface
}
type AuthenticatedUser struct {
	ID        string
	Login     string
	ExpiresAt time.Time
}
type SessionKey struct {
	ProviderID string
	SessionID  string
}

type Iface interface {
	GetAuthURL(provider Provider) (string, error)
	GetRouter() *http.ServeMux
	GetAuthenticatedUser(provider Provider, sessionID string) (AuthenticatedUser, error)
	HasValidMetadataURL(metadataURL string) (bool, error)
	CreateSession(key SessionKey, value AuthenticatedUser)
}

type AuthnRequest struct {
	XMLName                     xml.Name `xml:"AuthnRequest"`
	Text                        string   `xml:",chardata"`
	Samlp                       string   `xml:"samlp,attr"`
	Saml                        string   `xml:"saml,attr"`
	ID                          string   `xml:"ID,attr"`
	Version                     string   `xml:"Version,attr"`
	ProtocolBinding             string   `xml:"ProtocolBinding,attr"`
	AssertionConsumerServiceURL string   `xml:"AssertionConsumerServiceURL,attr"`
	IssueInstant                string   `xml:"IssueInstant,attr"`
	Destination                 string   `xml:"Destination,attr"`
	Issuer                      string   `xml:"Issuer"`
	Signature                   struct {
		Text       string `xml:",chardata"`
		Ds         string `xml:"ds,attr"`
		SignedInfo struct {
			Text                   string `xml:",chardata"`
			CanonicalizationMethod struct {
				Text      string `xml:",chardata"`
				Algorithm string `xml:"Algorithm,attr"`
			} `xml:"CanonicalizationMethod"`
			SignatureMethod struct {
				Text      string `xml:",chardata"`
				Algorithm string `xml:"Algorithm,attr"`
			} `xml:"SignatureMethod"`
			Reference struct {
				Text       string `xml:",chardata"`
				URI        string `xml:"URI,attr"`
				Transforms struct {
					Text      string `xml:",chardata"`
					Transform []struct {
						Text      string `xml:",chardata"`
						Algorithm string `xml:"Algorithm,attr"`
					} `xml:"Transform"`
				} `xml:"Transforms"`
				DigestMethod struct {
					Text      string `xml:",chardata"`
					Algorithm string `xml:"Algorithm,attr"`
				} `xml:"DigestMethod"`
				DigestValue string `xml:"DigestValue"`
			} `xml:"Reference"`
		} `xml:"SignedInfo"`
		SignatureValue string `xml:"SignatureValue"`
		KeyInfo        struct {
			Text     string `xml:",chardata"`
			X509Data struct {
				Text            string `xml:",chardata"`
				X509Certificate string `xml:"X509Certificate"`
			} `xml:"X509Data"`
		} `xml:"KeyInfo"`
	} `xml:"Signature"`
	NameIDPolicy struct {
		Text        string `xml:",chardata"`
		AllowCreate string `xml:"AllowCreate,attr"`
	} `xml:"NameIDPolicy"`
}

type Response struct {
	XMLName      xml.Name          `xml:"Response"`
	Text         string            `xml:",chardata"`
	Saml         string            `xml:"xmlns:saml,attr"`
	Samlp        string            `xml:"xmlns:samlp,attr"`
	ID           string            `xml:"ID,attr"`
	Version      string            `xml:"Version,attr"`
	IssueInstant string            `xml:"IssueInstant,attr"`
	Destination  string            `xml:"Destination,attr"`
	Issuer       string            `xml:"saml:Issuer"`
	Signature    ResponseSignature `xml:"ds:Signature"`
	Status       struct {
		Text       string `xml:",chardata"`
		StatusCode struct {
			Text  string `xml:",chardata"`
			Value string `xml:"Value,attr"`
		} `xml:"StatusCode"`
	} `xml:"Status"`
	Assertion ResponseAssertion `xml:"Assertion"`
}

type ResponseSignature struct {
	XMLName        xml.Name                    `xml:"ds:Signature"`
	Text           string                      `xml:",chardata"`
	Ds             string                      `xml:"xmlns:ds,attr"`
	SignedInfo     ResponseSignatureSignedInfo `xml:"ds:SignedInfo"`
	SignatureValue string                      `xml:"ds:SignatureValue"`
	KeyInfo        ResponseSignatureKeyInfo    `xml:"ds:KeyInfo"`
}

type ResponseSignatureSignedInfo struct {
	Text                   string `xml:",chardata"`
	CanonicalizationMethod struct {
		Text      string `xml:",chardata"`
		Algorithm string `xml:"Algorithm,attr"`
	} `xml:"ds:CanonicalizationMethod"`
	SignatureMethod ResponseSignatureSignedInfoSignatureMethod `xml:"ds:SignatureMethod"`
	Reference       ResponseSignatureSignedInfoReference       `xml:"ds:Reference"`
}
type ResponseSignatureSignedInfoSignatureMethod struct {
	Text      string `xml:",chardata"`
	Algorithm string `xml:"Algorithm,attr"`
}
type ResponseSignatureSignedInfoReference struct {
	Text       string `xml:",chardata"`
	URI        string `xml:"URI,attr"`
	Transforms struct {
		Text      string `xml:",chardata"`
		Transform []struct {
			Text      string `xml:",chardata"`
			Algorithm string `xml:"Algorithm,attr"`
		} `xml:"ds:Transform"`
	} `xml:"ds:Transforms"`
	DigestMethod struct {
		Text      string `xml:",chardata"`
		Algorithm string `xml:"Algorithm,attr"`
	} `xml:"ds:DigestMethod"`
	DigestValue string `xml:"ds:DigestValue"`
}
type ResponseSignatureKeyInfo struct {
	Text     string `xml:",chardata"`
	X509Data struct {
		Text            string `xml:",chardata"`
		X509Certificate string `xml:"ds:X509Certificate"`
	} `xml:"ds:X509Data"`
}

type ResponseConditions struct {
	Text                string                               `xml:",chardata"`
	NotBefore           string                               `xml:"NotBefore,attr"`
	NotOnOrAfter        string                               `xml:"NotOnOrAfter,attr"`
	AudienceRestriction ResponseConditionsAdienceRestriction `xml:"AudienceRestriction"`
}
type ResponseConditionsAdienceRestriction struct {
	Text     string `xml:",chardata"`
	Audience string `xml:"Audience"`
}
type ResponseSubject struct {
	Text   string `xml:",chardata"`
	NameID struct {
		Text   string `xml:",chardata"`
		Format string `xml:"Format,attr"`
	} `xml:"NameID"`
	SubjectConfirmation struct {
		Text                    string `xml:",chardata"`
		Method                  string `xml:"Method,attr"`
		SubjectConfirmationData struct {
			Text         string `xml:",chardata"`
			NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
			Recipient    string `xml:"Recipient,attr"`
		} `xml:"SubjectConfirmationData"`
	} `xml:"SubjectConfirmation"`
}

type ResponseAssertion struct {
	Text           string             `xml:",chardata"`
	Saml           string             `xml:"saml,attr"`
	Xs             string             `xml:"xs,attr"`
	Xsi            string             `xml:"xsi,attr"`
	Version        string             `xml:"Version,attr"`
	ID             string             `xml:"ID,attr"`
	IssueInstant   string             `xml:"IssueInstant,attr"`
	Issuer         string             `xml:"Issuer"`
	Subject        ResponseSubject    `xml:"Subject"`
	Conditions     ResponseConditions `xml:"Conditions"`
	AuthnStatement struct {
		Text                string `xml:",chardata"`
		AuthnInstant        string `xml:"AuthnInstant,attr"`
		SessionNotOnOrAfter string `xml:"SessionNotOnOrAfter,attr"`
		SessionIndex        string `xml:"SessionIndex,attr"`
		AuthnContext        struct {
			Text                 string `xml:",chardata"`
			AuthnContextClassRef string `xml:"AuthnContextClassRef"`
		} `xml:"AuthnContext"`
	} `xml:"AuthnStatement"`
}
