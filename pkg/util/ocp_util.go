package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"time"
)

// Generates a new self signed certificate
// and returns it along with the private key
func GenAdminKubeconfigSignerCert() ([]byte, []byte, error) {
	ca := &x509.Certificate{
		Subject: pkix.Name{
			OrganizationalUnit: []string{"openshift"},
			CommonName:         "admin-kubeconfig-signer-custom",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})

	return certPem, privateKeyPem, nil
}
