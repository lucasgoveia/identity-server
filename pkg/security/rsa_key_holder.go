package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type RSAKeyHolder struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func NewRSAKeyHolder(privateB64Key string, publicB64Key string) (*RSAKeyHolder, error) {
	privateKey, err := decodePrivateKeyFromBase64(privateB64Key)

	if err != nil {
		return nil, err
	}

	publicKey, err := decodePublicKeyFromBase64(publicB64Key)

	if err != nil {
		return nil, err
	}

	return &RSAKeyHolder{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func decodePrivateKeyFromBase64(base64Key string) (*rsa.PrivateKey, error) {
	privBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return privateKey, nil
}

func decodePublicKeyFromBase64(base64Key string) (*rsa.PublicKey, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}
	publicKey, err := x509.ParsePKCS1PublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	return publicKey, nil
}
