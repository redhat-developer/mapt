package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

const (
	PRIVATE_KEY_TYPE string = "RSA PRIVATE KEY"
	PUBLIC_KEY_TYPE  string = "RSA PUBLIC KEY"
)

func CreateKey(size int) (keyPEM, pubPEM []byte) {
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		panic(err)
	}
	pub := key.Public()
	keyPEM = pem.EncodeToMemory(
		&pem.Block{
			Type:  PRIVATE_KEY_TYPE,
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	pubPEM = pem.EncodeToMemory(
		&pem.Block{
			Type:  PUBLIC_KEY_TYPE,
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)
	return
}
