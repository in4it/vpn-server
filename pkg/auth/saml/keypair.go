package saml

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
)

type KeyPair struct {
	storage  storage.Iface
	hostname string
}

func NewKeyPair(storage storage.Iface, hostname string) *KeyPair {
	return &KeyPair{
		storage:  storage,
		hostname: hostname,
	}
}

func (kp *KeyPair) GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error) {
	if !kp.storage.FileExists(kp.storage.ConfigPath("saml/saml.key")) {
		err := kp.generateKeyAndCert()
		if err != nil {
			return privateKey, cert, fmt.Errorf("can't generate saml key and cert: %s", err)
		}
	} else if !kp.storage.FileExists(kp.storage.ConfigPath("saml/saml.crt")) {
		err = kp.generateKeyAndCert()
		if err != nil {
			return privateKey, cert, fmt.Errorf("can't generate saml key and cert: %s", err)
		}
	}

	certPEMBlock, err := kp.storage.ReadFile(kp.storage.ConfigPath("saml/saml.crt"))
	if err != nil {
		return privateKey, cert, fmt.Errorf("can't read saml certificate: %s", err)
	}
	keyPEMBlock, err := kp.storage.ReadFile(kp.storage.ConfigPath("saml/saml.key"))
	if err != nil {
		return privateKey, cert, fmt.Errorf("can't read saml key: %s", err)
	}
	keyPair, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return privateKey, cert, fmt.Errorf("can't get saml keypair: %s", err)
	}

	privateKey = keyPair.PrivateKey.(*rsa.PrivateKey)
	cert = keyPair.Certificate[0]

	return privateKey, cert, nil
}

func (kp *KeyPair) generateKeyAndCert() error {
	var (
		certOut bytes.Buffer
		keyOut  bytes.Buffer
	)

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("private key generation failed: %s", err)
	}

	if err = pem.Encode(&keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("rand int error: %s", err)
	}
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: kp.hostname,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),

		IsCA: true,

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("certificate creation error: %s", err)
	}
	if err = pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	// ensure storagepath exists
	err = kp.storage.EnsurePath(kp.storage.ConfigPath("saml"))
	if err != nil {
		return fmt.Errorf("could not ensure saml path exists: %s", err)
	}

	err = kp.storage.WriteFile(kp.storage.ConfigPath("saml/saml.key"), keyOut.Bytes())
	if err != nil {
		return fmt.Errorf("saml key write error: %s", err)
	}
	err = kp.storage.WriteFile(kp.storage.ConfigPath("saml/saml.crt"), certOut.Bytes())
	if err != nil {
		return fmt.Errorf("saml key write error: %s", err)
	}

	return nil
}
