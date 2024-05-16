package rest

/*
 * Genarate rsa keys. (https://github.com/wardviaene/http-echo/blob/master/rsa.go)
 */

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path"

	"github.com/golang-jwt/jwt/v5"
	"github.com/in4it/wireguard-server/pkg/storage"
)

type JWTKeys struct {
	PrivateKey *rsa.PrivateKey `json:"privateKey,omitempty"`
	PublicKey  *rsa.PublicKey  `json:"publicKey,omitempty"`
}

func getJWTKeys(storage storage.Iface) (*JWTKeys, error) {

	filename := storage.ConfigPath("pki/private.pem")
	filenamePublicKey := storage.ConfigPath("pki/public.pem")

	if !storage.FileExists(filename) {
		err := storage.EnsurePath(path.Dir(filename))
		if err != nil {
			return nil, fmt.Errorf("ensure path error: %s", err)
		}
		err = createJWTKeys(storage, storage.ConfigPath("pki"))
		if err != nil {
			return nil, fmt.Errorf("createJWTKeys error: %s", err)
		}
	}

	signBytes, err := storage.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("private key read error: %s", err)
	}
	publicBytes, err := storage.ReadFile(filenamePublicKey)
	if err != nil {
		return nil, fmt.Errorf("private key read error: %s", err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse private key: %s", err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		return nil, fmt.Errorf("can't parse public key: %s", err)
	}
	return &JWTKeys{PrivateKey: signKey, PublicKey: publicKey}, nil
}

func createJWTKeys(storage storage.Iface, path string) error {
	reader := rand.Reader
	bitSize := 4096

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return err
	}

	publicKey := key.PublicKey

	err = savePEMKey(storage, path+"/private.pem", key)
	if err != nil {
		return err
	}
	err = savePublicPEMKey(storage, path+"/public.pem", publicKey)
	if err != nil {
		return err
	}

	return nil
}

func savePEMKey(storage storage.Iface, fileName string, key *rsa.PrivateKey) error {
	var buf bytes.Buffer

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err := pem.Encode(&buf, privateKey)
	if err != nil {
		return err
	}

	err = storage.WriteFile(fileName, buf.Bytes())
	if err != nil {
		return fmt.Errorf("WriteFile error: %s", err)
	}
	return nil
}

func savePublicPEMKey(storage storage.Iface, fileName string, pubkey rsa.PublicKey) error {
	var buf bytes.Buffer

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&pubkey)
	if err != nil {
		return err
	}

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	err = pem.Encode(&buf, pemkey)
	if err != nil {
		return err
	}
	err = storage.WriteFile(fileName, buf.Bytes())
	if err != nil {
		return fmt.Errorf("WriteFile error: %s", err)
	}
	return nil
}
