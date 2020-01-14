package orbiter

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

type EnsureFunc func(nodeAgentsCurrent map[string]*NodeAgentCurrent, nodeAgentsDesired map[string]*NodeAgentSpec) (err error)

type AdaptFunc func(desired *Tree, secrets *Tree, current *Tree) (EnsureFunc, error)

type Common struct {
	Kind    string
	Version string
}

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = *node
	err := node.Decode(&c.Common)
	return err
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}

type Secret struct {
	Encryption string
	Encoding   string
	Value      string
	Masterkey  string `yaml:"-"`
}

func (s *Secret) UnmarshalYAML(node *yaml.Node) error {

	type Alias Secret
	alias := &Secret{}
	if err := node.Decode(alias); err != nil {
		return err
	}
	s.Encryption = alias.Encryption
	s.Encoding = alias.Encoding
	if alias.Value == "" {
		return nil
	}

	cipherText, err := base64.URLEncoding.DecodeString(alias.Value)
	if err != nil {
		return err
	}

	if len(s.Masterkey) < 1 || len(s.Masterkey) > 32 {
		return errors.New("Master key size must be between 1 and 32 characters")
	}

	masterKey := make([]byte, 32)
	for idx, char := range []byte(strings.Trim(s.Masterkey, "\n")) {
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
	s.Value = string(cipherText)
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

	type Alias Secret
	return &Alias{Encryption: "AES256", Encoding: "Base64", Value: base64.URLEncoding.EncodeToString(cipherText)}, nil
}

func Iterator(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, destroy bool, adapt AdaptFunc) func() {

	return func() {
		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		rawDesired, err := gitClient.Read("desired.yml")
		if err != nil {
			logger.Error(err)
			return
		}
		treeDesired := &Tree{}
		if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
			panic(err)
		}

		rawSecrets, err := gitClient.Read("secrets.yml")
		if err != nil {
			logger.Error(err)
			return
		}
		treeSecrets := &Tree{}
		if err := yaml.Unmarshal([]byte(rawSecrets), treeSecrets); err != nil {
			panic(err)
		}

		treeCurrent := &Tree{}
		ensure, err := adapt(treeDesired, treeSecrets, treeCurrent)
		if err != nil {
			logger.Error(err)
			return
		}

		desiredNodeAgents := make(map[string]*NodeAgentSpec)
		currentNodeAgents := NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, _ := gitClient.Read("internal/node-agents-current.yml")
		if rawCurrentNodeAgents != nil {
			yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)
		}

		if err := ensure(currentNodeAgents.Current, desiredNodeAgents); err != nil {
			logger.Error(err)
			return
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeDesired)
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "secrets.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeSecrets)
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "internal/node-agents-desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(&NodeAgentsDesiredKind{
					Common: Common{
						Kind:    "nodeagent.caos.ch/NodeAgent",
						Version: "v0",
					},
					Spec: NodeAgentsSpec{
						Commit:     orbiterCommit,
						NodeAgents: desiredNodeAgents,
					},
				})
			}}); err != nil {
			panic(err)
		}

		newCurrent, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "current.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeCurrent)
			}})

		if err != nil {
			panic(err)
		}

		statusReader := struct {
			Deps struct {
				Clusters map[string]struct {
					Current struct {
						State struct {
							Status string
						}
					}
				}
			}
		}{}
		if err := yaml.Unmarshal(newCurrent, &statusReader); err != nil {
			panic(err)
		}
		for _, cluster := range statusReader.Deps.Clusters {
			if destroy && cluster.Current.State.Status != "destroyed" ||
				!destroy && !recur && cluster.Current.State.Status == "running" {
				os.Exit(0)
			}
		}
	}

}
