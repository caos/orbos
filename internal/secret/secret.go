package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/caos/orbos/internal/utils/clientgo"
	"gopkg.in/yaml.v3"
)

var Masterkey = "empty"

// Secret: Secret handled with orbctl so no manual changes are required
type Secret struct {
	//Encryption algorithm used for the secret
	Encryption string
	//Encoding algorithm used for the secret
	Encoding string
	//Encrypted and encoded Value
	Value string
}
type secretAlias Secret

// Existing: Used secret that has to be already existing in the cluster
type Existing struct {
	//Name of the Secret
	Name string `json:"name" yaml:"name"`
	//Key in the secret from where the value should be used
	Key string `json:"key" yaml:"key"`
	//Name which should be used internally, should be unique for the volume and volumemounts
	InternalName string `json:"internalName,omitempty" yaml:"internalName,omitempty"`
}

// Existing: Used secret that has to be already existing in the cluster and should contain id/username and secret/password
type ExistingIDSecret struct {
	//Name of the Secret
	Name string `json:"name" yaml:"name"`
	//Key in the secret which contains the ID
	IDKey string `json:"idKey" yaml:"idKey"`
	//Key in the secret which contains the secret
	SecretKey string `json:"secretKey" yaml:"secretKey"`
	//Name which should be used internally, should be unique for the volume and volumemounts
	InternalName string `json:"internalName,omitempty" yaml:"internalName,omitempty"`
}

func (s *Secret) UnmarshalYAMLWithExisting(node *yaml.Node, existing *Existing) error {
	if err := s.UnmarshalYAML(node); err != nil {
		return err
	}

	if s.Value == "" {
		if existing != nil && existing.Name != "" && existing.Key != "" {
			secret, err := clientgo.GetSecret(existing.Name, "caos-system")
			if err != nil {
				return errors.New("Error while reading existing secret")
			}

			value, found := secret.Data[existing.Key]
			if !found {
				return errors.New("Error while reading existing secret, key non-existent")
			}
			s.Value = string(value)
		}
	}

	return nil
}

func unmarshal(s *Secret) (string, error) {
	if s.Value == "" {
		return "", nil
	}

	cipherText, err := base64.URLEncoding.DecodeString(s.Value)
	if err != nil {
		return "", err
	}

	if len(Masterkey) < 1 || len(Masterkey) > 32 {
		return "", nil
		//return errors.New("Master key size must be between 1 and 32 characters")
	}

	masterKeyLocal := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(Masterkey, "\n")) {
		masterKeyLocal[idx] = char
	}

	block, err := aes.NewCipher(masterKeyLocal)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("Ciphertext block size is too short")
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	if !utf8.Valid(cipherText) {
		return "", errors.New("Decryption failed")
	}
	//	s.monitor.Info("Decoded and decrypted secret")
	return string(cipherText), nil
}

func (s *Secret) Unmarshal(masterkey string) error {
	Masterkey = masterkey

	unm, err := unmarshal(s)
	if err != nil {
		return err
	}

	s.Value = unm
	return nil
}

func (s *Secret) UnmarshalYAML(node *yaml.Node) error {
	alias := new(secretAlias)
	err := node.Decode(alias)

	if alias.Value == "" {
		return nil
	}

	s.Encoding = alias.Encoding
	s.Encryption = alias.Encryption
	s.Value = alias.Value

	if len(Masterkey) < 1 || len(Masterkey) > 32 {
		return nil
		//return errors.New("Master key size must be between 1 and 32 characters")
	}

	unmarshalled, err := unmarshal(s)
	if err != nil {
		return err
	}

	//	s.monitor.Info("Decoded and decrypted secret")
	s.Encoding = alias.Encoding
	s.Encryption = alias.Encryption
	s.Value = unmarshalled
	return nil
}

func (s *Secret) MarshalYAML() (interface{}, error) {

	if s.Value == "" {
		return nil, nil
	}

	if len(Masterkey) < 1 || len(Masterkey) > 32 {
		return nil, errors.New("Master key size must be between 1 and 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(Masterkey, "\n")) {
		masterKey[idx] = char
	}

	c, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(s.Value))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(c, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(s.Value))

	return &secretAlias{Encryption: "AES256", Encoding: "Base64", Value: base64.URLEncoding.EncodeToString(cipherText)}, nil
}

func InitIfNil(sec *Secret) *Secret {
	if sec == nil {
		return &Secret{}
	}
	return sec
}
