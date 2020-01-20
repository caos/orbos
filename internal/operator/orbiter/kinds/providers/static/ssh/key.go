package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	sshlib "golang.org/x/crypto/ssh"
)

func Generate() (private string, public string, err error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return private, public, err
	}

	if err := privateKey.Validate(); err != nil {
		return private, public, err
	}

	publicKey, err := sshlib.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return private, public, err
	}

	enc := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return strings.TrimSpace(string(enc)), string(sshlib.MarshalAuthorizedKey(publicKey)), nil
}
