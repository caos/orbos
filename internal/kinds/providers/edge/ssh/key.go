package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

func EnsureKeyPair(sec *orbiter.Secrets, keyProperty string, pubKeyProperty string) ([]byte, error) {

	existing, err := sec.Read(pubKeyProperty)
	if err != nil && errors.Cause(err) != secret.ErrNotExist {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	privateKey, publicKey, err := generate()
	if err != nil {
		return nil, err
	}

	if err := sec.Write(keyProperty, privateKey); err != nil {
		return nil, err
	}
	if err := sec.Write(pubKeyProperty, publicKey); err != nil {
		return nil, err
	}

	return publicKey, nil
}

func generate() ([]byte, []byte, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	err = privateKey.Validate()
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := sshlib.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	enc := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return []byte(strings.TrimSpace(string(enc))), sshlib.MarshalAuthorizedKey(publicKey), nil
}
