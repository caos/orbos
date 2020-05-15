package ssh

import (
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

type pair struct {
	private []byte
	public  sshlib.AuthMethod
}

var (
	cachedKeys []pair
)

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
