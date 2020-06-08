package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
	"strings"
)

type pair struct {
	private []byte
	public  sshlib.AuthMethod
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

func PrivateKeyToPublicKey(privKey []byte) (sshlib.AuthMethod, error) {
	cached := false
	var pubKey sshlib.AuthMethod
	for _, cachedKey := range cachedKeys {
		if string(cachedKey.private) == string(privKey) {
			pubKey = cachedKey.public
			cached = true
		}
	}
	if !cached {
		signer, err := sshlib.ParsePrivateKey(privKey)
		if err != nil {
			return nil, errors.Wrap(err, "parsing private key failed")
		}
		pubKey = sshlib.PublicKeys(signer)
		cachedKeys = append(cachedKeys, pair{privKey, pubKey})
	}

	return pubKey, nil
}
