package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
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
	file      map[string]interface{}
	property  string
	masterKey string
}

func New(logger logging.Logger, file map[string]interface{}, property string, masterKey string) *Secret {
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
func (s *Secret) Write(secret []byte) (err error) {

	defer func() {
		err = errors.Wrapf(err, "Writing secret %s failed", s.property)
	}()

	if !utf8.Valid(secret) {
		return errors.Errorf("Binary secrets are not allowed")
	}

	if len(s.masterKey) > 32 {
		return errors.New("Master key cannot be longer than 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.masterKey, "\n")) {
		masterKey[idx] = char
	}

	c, err := aes.NewCipher(masterKey)
	if err != nil {
		return err
	}

	cipherText := make([]byte, aes.BlockSize+len(secret))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(c, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], secret)

	s.file[s.property] = &Property{Encryption: "AES256", Encoding: "Base64", Value: base64.URLEncoding.EncodeToString(cipherText)}
	s.logger.Info("Encrypted and encoded secret")
	return nil
}

func (s *Secret) Read(to io.Writer) (err error) {

	defer func() {
		err = errors.Wrapf(err, "Reading secret %s failed", s.property)
	}()

	properties := make(map[string]Property)
	if err := mapstructure.Decode(s.file, &properties); err != nil {
		return err
	}

	property, ok := properties[s.property]
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
