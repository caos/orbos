package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"time"
)

func NewCA() (*rsa.PrivateKey, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"Cockroach"},
			CommonName:   "Cockroach CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return caPrivKey, caBytes, nil
}

func NewClient(caPrivKey *rsa.PrivateKey, ca []byte, user string) (*rsa.PrivateKey, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Cockroach"},
			CommonName:   user,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := x509.ParseCertificate(ca)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return certPrivKey, certBytes, nil
}

func NewNode(caPrivKey *rsa.PrivateKey, ca []byte, namespace string, clusterDns string) (*rsa.PrivateKey, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Cockroach"},
			CommonName:   "node",
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames: []string{
			//TODO clustername
			"localhost",
			"cockroachdb-public",
			"cockroachdb-public.default",
			"cockroachdb-public." + namespace,
			"cockroachdb-public." + namespace + ".svc." + clusterDns,
			"*.cockroachdb",
			"*.cockroachdb." + namespace,
			"*.cockroachdb." + namespace + ".svc." + clusterDns,
		},
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := x509.ParseCertificate(ca)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return certPrivKey, certBytes, nil
}

/*
func PEMEncodeCertificate(certData []byte) ([]byte, error) {
	data, err := x509.EncryptPEMBlock(rand.Reader, "CERTIFICATE", certData, []byte(""), x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}

	keyPem := new(bytes.Buffer)
	if err := pem.Encode(keyPem, data); err != nil {
		return nil, err
	}
	return keyPem.Bytes(), nil
}

func PEMEncodeKey(key *rsa.PrivateKey) ([]byte, error) {
	keyData := x509.MarshalPKCS1PrivateKey(key)
	data, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", keyData, []byte(""), x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}
	keyPem := new(bytes.Buffer)
	if err := pem.Encode(keyPem, data); err != nil {
		return nil, err
	}
	return keyPem.Bytes(), nil
}

func PEMDecodeKey(data []byte) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	if len(rest) > 0 {
		return nil, errors.New("extra data")
	}
	der, err := x509.DecryptPEMBlock(block, []byte(""))
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(der)
}
*/

func PEMEncodeCertificate(data []byte) ([]byte, error) {
	certPem := new(bytes.Buffer)
	if err := pem.Encode(certPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: data,
	}); err != nil {
		return nil, err
	}
	return certPem.Bytes(), nil
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

func PEMDecodeKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
