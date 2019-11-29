package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/core/logging"
)

type constErr string

func (c constErr) Error() string { return string(c) }

const ErrNotExist = constErr("Secret does not exist")

type Property struct {
	Encryption string
	Encoding   string
	Value      string
}

type Secret struct {
	logger    logging.Logger
	file      []byte
	property  string
	masterKey string
}

func New(logger logging.Logger, file []byte, property string, masterKey string) *Secret {
	return &Secret{
		logger: logger.WithFields(map[string]interface{}{
			"property": property,
		}),
		file:      file,
		property:  property,
		masterKey: masterKey,
	}
}

// Write returns the whole file passed in secret.New() updated at the property passed in secret.New()
func (s *Secret) Write(secret []byte) (newFile []byte, err error) {

	defer func() {
		err = errors.Wrapf(err, "Writing secret %s failed", s.property)
	}()

	if !utf8.Valid(secret) {
		return nil, errors.Errorf("Binary secrets are not allowed")
	}

	if len(s.masterKey) > 32 {
		return nil, errors.New("Master key cannot be longer than 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.masterKey, "\n")) {
		masterKey[idx] = char
	}

	c, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(secret))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(c, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], secret)

	secrets := make(map[string]*Property)
	if err := yaml.Unmarshal(s.file, secrets); err != nil {
		return nil, err
	}

	secrets[s.property] = &Property{Encryption: "AES256", Encoding: "Base64", Value: base64.URLEncoding.EncodeToString(cipherText)}
	newSecrets, err := yaml.Marshal(secrets)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Encrypted and encoded secret")
	return newSecrets, nil
}

func (s *Secret) Read(to io.Writer) (err error) {

	defer func() {
		err = errors.Wrapf(err, "Reading secret %s failed", s.property)
	}()

	secrets := make(map[string]*Property)
	if err := yaml.Unmarshal(s.file, secrets); err != nil {
		return err
	}

	property, ok := secrets[s.property]
	if !ok {
		return ErrNotExist
	}

	cipherText, err := base64.URLEncoding.DecodeString(property.Value)
	if err != nil {
		return err
	}

	if len(s.masterKey) > 32 {
		return errors.New("Master key cannot be longer than 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.masterKey, "\n")) {
		masterKey[idx] = char
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return err
	}

	if len(cipherText) < aes.BlockSize {
		return errors.New("Ciphertext block size is too short")
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	if !utf8.Valid(cipherText) {
		return errors.New("Decryption failed")
	}
	//	s.logger.Info("Decoded and decrypted secret")

	_, err = to.Write(cipherText)
	return err
}
