package operator

import (
	"bytes"
	"sync"

	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/core/secret"
	"gopkg.in/yaml.v2"
)

type curriedSecrets struct {
	logger           logging.Logger
	encryptedSecrets map[string]interface{}
	masterkey        string
	//	changes          []func(map[string]interface{}) error
	mux          sync.Mutex
	updateRemote func([]byte) error
}

func currySecrets(logger logging.Logger, updateRemote func([]byte) error, secrets map[string]interface{}, masterkey string) *curriedSecrets {
	return &curriedSecrets{
		logger:           logger,
		encryptedSecrets: secrets,
		masterkey:        masterkey,
		updateRemote:     updateRemote,
		//		changes:          make([]func(map[string]interface{}) error, 0),
	}
}

func (c *curriedSecrets) read(property string) ([]byte, error) {
	serialized, err := yaml.Marshal(c.encryptedSecrets)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := secret.New(c.logger, serialized, property, c.masterkey).Read(&buf); err != nil {
		return nil, err
	}
	c.logger.Debug("Secret read")
	return buf.Bytes(), nil
}
func (c *curriedSecrets) write(property string, value []byte) error {

	serialized, err := yaml.Marshal(c.encryptedSecrets)
	if err != nil {
		return err
	}

	newSecrets, err := secret.New(c.logger, serialized, property, c.masterkey).Write(value)
	if err != nil {
		return err
	}
	if err := c.updateRemote(newSecrets); err != nil {
		return err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return yaml.Unmarshal(newSecrets, c.encryptedSecrets)
}

func (c *curriedSecrets) delete(property string) error {
	delete(c.encryptedSecrets, property)
	serialized, err := yaml.Marshal(c.encryptedSecrets)
	if err != nil {
		return err
	}
	return c.updateRemote(serialized)
}
