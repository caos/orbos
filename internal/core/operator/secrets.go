package operator

import (
	"bytes"
	"sync"

	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/logging"
)

type curriedSecrets struct {
	logger           logging.Logger
	encryptedSecrets map[string]interface{}
	masterkey        string
	//	changes          []func(map[string]interface{}) error
	mux          sync.Mutex
	updateRemote func(map[string]interface{}) error
}

func currySecrets(logger logging.Logger, updateRemote func(map[string]interface{}) error, secrets map[string]interface{}, masterkey string) *curriedSecrets {
	return &curriedSecrets{
		logger:           logger,
		encryptedSecrets: secrets,
		masterkey:        masterkey,
		updateRemote:     updateRemote,
		//		changes:          make([]func(map[string]interface{}) error, 0),
	}
}

func (c *curriedSecrets) read(property string) ([]byte, error) {
	var buf bytes.Buffer
	if err := secret.New(c.logger, c.encryptedSecrets, property, c.masterkey).Read(&buf); err != nil {
		return nil, err
	}
	c.logger.Debug("Secret read")
	return buf.Bytes(), nil
}
func (c *curriedSecrets) write(property string, value []byte) error {
	err := secret.New(c.logger, c.encryptedSecrets, property, c.masterkey).Write(value)
	if err != nil {
		return err
	}
	return c.updateRemote(c.encryptedSecrets)
}

func (c *curriedSecrets) delete(property string) error {
	delete(c.encryptedSecrets, property)
	return c.updateRemote(c.encryptedSecrets)
}
