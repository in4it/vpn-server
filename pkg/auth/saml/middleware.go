package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

const ISSUER_URL = "saml/iss"
const AUDIENCE_URL = "saml/aud"
const ACS_URL = "saml/acs"

func (s *saml) ensureSPLoaded(provider Provider) error {
	if _, ok := s.serviceProvider[provider.ID]; !ok {
		err := s.loadSP(provider)
		if err != nil {
			return fmt.Errorf("could not load saml provider: %s", err)
		}
	} else {
		// check if provider is up-to-date
		if provider.AllowMissingAttributes != s.serviceProvider[provider.ID].AllowMissingAttributes {
			s.serviceProvider[provider.ID] = nil
		}
		if s.serviceProvider[provider.ID] == nil {
			err := s.loadSP(provider)
			if err != nil {
				return fmt.Errorf("could not reload saml provider: %s", err)
			}
		}
	}
	return nil
}

func (s *saml) loadSP(provider Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idpMetadataURL, err := url.Parse(provider.MetadataURL)
	if err != nil {
		return fmt.Errorf("can't parse metadata url: %s", err)
	}
	// pull metadata
	res, err := http.Get(idpMetadataURL.String())
	if err != nil {
		return fmt.Errorf("can't retrieve saml metadata: %s", err)
	}

	rawMetadata, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("can't read saml cert data: %s", err)
	}

	metadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return fmt.Errorf("can't decode saml cert data: %s", err)
	}

	// load certs
	certStore := dsig.MemoryX509CertificateStore{}

	if metadata.IDPSSODescriptor == nil || len(metadata.IDPSSODescriptor.KeyDescriptors) == 0 {
		return fmt.Errorf("keyDescriptors are empty")
	}
	if len(metadata.IDPSSODescriptor.SingleSignOnServices) == 0 {
		return fmt.Errorf("SingleSignOnServices not found")
	}

	certStore.Roots, err = getSAMLCertsFromMetadata(metadata.IDPSSODescriptor.KeyDescriptors)
	if err != nil {
		return fmt.Errorf("can't parse certs from metadata: %s", err)
	}

	keyStore := NewKeyPair(s.storage, *s.hostname)

	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       fmt.Sprintf("%s://%s/%s/%s", *s.protocol, *s.hostname, ISSUER_URL, provider.ID),
		AssertionConsumerServiceURL: fmt.Sprintf("%s://%s/%s/%s", *s.protocol, *s.hostname, ACS_URL, provider.ID),
		SignAuthnRequests:           true,
		AudienceURI:                 fmt.Sprintf("%s://%s/%s/%s", *s.protocol, *s.hostname, AUDIENCE_URL, provider.ID),
		IDPCertificateStore:         &certStore,
		SPKeyStore:                  keyStore,
		AllowMissingAttributes:      provider.AllowMissingAttributes,
	}

	s.serviceProvider[provider.ID] = sp

	return err
}

func getSAMLCertsFromMetadata(keyDescriptors []types.KeyDescriptor) ([]*x509.Certificate, error) {
	certs := []*x509.Certificate{}

	for _, kd := range keyDescriptors {
		for idx, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				return nil, fmt.Errorf("metadata certificate(%d) must not be empty", idx)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return nil, fmt.Errorf("decode error:%s", err)
			}

			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return nil, fmt.Errorf("cert parse error: %s", err)
			}

			certs = append(certs, idpCert)
		}
	}

	return certs, nil

}

func (s *saml) GetAuthURL(provider Provider) (string, error) {
	err := s.ensureSPLoaded(provider)
	if err != nil {
		return "", fmt.Errorf("saml error: invalid saml configuration: %s", err)
	}
	if _, ok := s.serviceProvider[provider.ID]; !ok {
		return "", fmt.Errorf("provider not found")
	}
	return s.serviceProvider[provider.ID].BuildAuthURL("")
}
