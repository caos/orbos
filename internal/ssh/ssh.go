package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

type pair struct {
	private []byte
	signer  sshlib.Signer
}

var (
	cachedKeys []pair
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

func AuthMethodFromKeys(privKey ...[]byte) (method sshlib.AuthMethod, err error) {

	var signers []sshlib.Signer
	for _, key := range privKey {
		for _, cachedKey := range cachedKeys {
			if string(cachedKey.private) == string(key) {
				signers = append(signers, cachedKey.signer)
				break
			}
		}
		signer, err := sshlib.ParsePrivateKey(key)
		if err != nil {
			return nil, errors.Wrap(err, "parsing private key failed")
		}
		cachedKeys = append(cachedKeys, pair{key, signer})
		signers = append(signers, signer)
	}

	return sshlib.PublicKeys(signers...), nil
}
