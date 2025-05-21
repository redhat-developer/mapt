package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func GenAdminKubeconfigSignerCert() string {
	caCertFileName := "custom-ca.crt"
	caKeyFileName := "custom-ca.key"

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
		logging.Error(err)
		return ""
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		logging.Error(err)
		return ""
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})

	_ = os.Remove(caCertFileName)
	_ = os.Remove(caKeyFileName)

	if err := os.WriteFile(caCertFileName, certPem, 0444); err != nil {
		logging.Error(err)
		return ""
	}
	if err := os.WriteFile(caKeyFileName, privateKeyPem, 0444); err != nil {
		logging.Error(err)
		return ""
	}

	return filepath.Join(".", caCertFileName)
}
