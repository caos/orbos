package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

func CreatePrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 4096)
}

func GetCSR(domain string, privKey *rsa.PrivateKey) ([]byte, error) {

	temp := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"CH"},
			Organization: []string{"caos AG"},
			Locality:     []string{"St. Gallen"},
			CommonName:   domain,
		},
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, temp, privKey)
	if err != nil {
		return nil, err
	}

	return csr, nil
}

func PEMEncodeKey(key *rsa.PrivateKey) ([]byte, error) {
	keyPem := new(bytes.Buffer)
	if err := pem.Encode(keyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, err
	}
	return keyPem.Bytes(), nil
}

func PEMEncodeCSR(data []byte) ([]byte, error) {
	certPem := new(bytes.Buffer)
	if err := pem.Encode(certPem, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: data,
	}); err != nil {
		return nil, err
	}
	return certPem.Bytes(), nil
}
