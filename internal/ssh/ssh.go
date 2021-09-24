package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	sshlib "golang.org/x/crypto/ssh"
)

type pair struct {
	private []byte
	signer  sshlib.Signer
}

var (
	cachedKeys []pair
)

func Generate() (private string, public string) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	if err := privateKey.Validate(); err != nil {
		panic(err)
	}

	publicKey, err := sshlib.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	enc := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return strings.TrimSpace(string(enc)), string(sshlib.MarshalAuthorizedKey(publicKey))
}

func ParsePrivateKeys(privKey ...[]byte) ([]sshlib.Signer, error) {

	var signers []sshlib.Signer
	for _, copyKey := range privKey {
		key := copyKey
		cached := false
		for _, cachedKey := range cachedKeys {
			if string(cachedKey.private) == string(key) {
				cached = true
				signers = append(signers, cachedKey.signer)
				break
			}
		}
		if !cached {
			signer, err := sshlib.ParsePrivateKey(key)
			if err != nil {
				return nil, fmt.Errorf("parsing private key failed: %w", err)
			}
			cachedKeys = append(cachedKeys, pair{key, signer})
			signers = append(signers, signer)
		}
	}

	return signers, nil
}
