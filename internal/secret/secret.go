package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/caos/orbos/internal/utils/clientgo"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
	"unicode/utf8"
)

type Secret struct {
	Encryption string
	Encoding   string
	Value      string
	Masterkey  string `yaml:"-"`
}
type secretAlias Secret

type Existing struct {
	Name         string `json:"name" yaml:"name"`
	Key          string `json:"key" yaml:"key"`
	InternalName string `json:"internalName" yaml:"internalName"`
}

type ExistingIDSecret struct {
	Name         string `json:"name" yaml:"name"`
	IDKey        string `json:"idKey" yaml:"idKey"`
	SecretKey    string `json:"secretKey" yaml:"secretKey"`
	InternalName string `json:"internalName" yaml:"internalName"`
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

	if len(s.Masterkey) < 1 || len(s.Masterkey) > 32 {
		return "", nil
		//return errors.New("Master key size must be between 1 and 32 characters")
	}

	masterKeyLocal := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.Masterkey, "\n")) {
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
	s.Masterkey = masterkey

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

	if len(s.Masterkey) < 1 || len(s.Masterkey) > 32 {
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

	if len(s.Masterkey) < 1 || len(s.Masterkey) > 32 {
		return nil, errors.New("Master key size must be between 1 and 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.Masterkey, "\n")) {
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

	return &secretAlias{Encryption: "AES256", Encoding: "Base64", Value: base64.URLEncoding.EncodeToString(cipherText), Masterkey: ""}, nil
}

func ClearEmpty(secret *Secret) *Secret {
	if secret != nil && secret.Value == "" {
		return nil
	}
	return secret
}

func InitIfNil(sec *Secret, masterkey string) *Secret {
	if sec == nil {
		return &Secret{Masterkey: masterkey}
	}
	sec.Masterkey = masterkey
	return sec
}
