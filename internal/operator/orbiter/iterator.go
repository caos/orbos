package orbiter

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
)

type AdaptFunc func(desired *Tree, secrets *Tree, current *Tree) (EnsureFunc, ReadSecretFunc, WriteSecretFunc, error)

type EnsureFunc func(psf PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error)

type PushSecretsFunc func() error

type ReadSecretFunc func(path []string) (string, error)

type WriteSecretFunc func(path []string, value string) error

func Takeoff(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, destroy bool, adapt AdaptFunc) func() {

	return func() {

		treeDesired, treeSecrets, err := parse(gitClient)
		if err != nil {
			logger.Error(err)
			return
		}

		treeCurrent := &Tree{}
		ensure, _, _, err := adapt(treeDesired, treeSecrets, treeCurrent)
		if err != nil {
			logger.Error(err)
			return
		}

		desiredNodeAgents := make(map[string]*common.NodeAgentSpec)
		currentNodeAgents := common.NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, _ := gitClient.Read("internal/node-agents-current.yml")
		yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)

		if err := ensure(writeSecretsFunc(gitClient, treeSecrets), currentNodeAgents.Current, desiredNodeAgents); err != nil {
			logger.Error(err)
			return
		}

		current := common.MarshalYAML(treeCurrent)

		if err := gitClient.UpdateRemote(git.File{
			Path:    "desired.yml",
			Content: common.MarshalYAML(treeDesired),
		}, git.File{
			Path:    "current.yml",
			Content: current,
		}, git.File{
			Path: "internal/node-agents-desired.yml",
			Content: common.MarshalYAML(&common.NodeAgentsDesiredKind{
				Kind:    "nodeagent.caos.ch/NodeAgent",
				Version: "v0",
				Spec: common.NodeAgentsSpec{
					Commit:     orbiterCommit,
					NodeAgents: desiredNodeAgents,
				},
			}),
		}); err != nil {
			panic(err)
		}

		statusReader := struct {
			Deps map[string]struct {
				Current struct {
					State struct {
						Status string
					}
				}
			}
		}{}
		if err := yaml.Unmarshal(current, &statusReader); err != nil {
			panic(err)
		}
		for _, cluster := range statusReader.Deps {
			if destroy && cluster.Current.State.Status == "destroyed" ||
				!destroy && !recur && cluster.Current.State.Status == "running" {
				os.Exit(0)
			}
		}
	}
}

func AdaptReadSecret(path []string, deps map[string]ReadSecretFunc, mapping map[string]*Secret) (string, error) {

	if len(path) == 0 {
		return "", errors.New("no path provided")
	}

	key := path[0]

	if len(path) == 1 {
		if len(mapping) == 0 {
			return "", errors.New("kind does not need or support secrets")
		}

		value, ok := mapping[key]
		if !ok {
			return "", errors.Errorf("unknown secret %s", key)
		}
		return value.Value, nil
	}

	if len(deps) == 0 {
		return "", errors.Errorf("kind does not need or support dependencies")
	}

	read, ok := deps[key]
	if !ok {
		return "", errors.Errorf("dependency %s not found", key)
	}

	if read == nil {
		return "", errors.Errorf("dependency %s does not need or support secrets", key)
	}

	val, err := read(path[1:])
	return val, errors.Wrapf(err, "reading secret from dependency %s failed", key)
}

func AdaptWriteSecret(path []string, value string, deps map[string]WriteSecretFunc, mapping map[string]*Secret) error {

	if len(path) == 0 {
		return errors.New("no path provided")
	}

	key := path[0]

	if len(path) == 0 {
		if len(mapping) == 0 {
			return errors.New("kind does not need or support secrets")
		}

		secret, ok := mapping[key]
		if !ok {
			return errors.Errorf("unknown secret %s", key)
		}
		secret.Value = value
		return nil
	}

	if len(deps) == 0 {
		return errors.Errorf("kind does not need or support dependencies")
	}

	write, ok := deps[key]
	if !ok {
		return errors.Errorf("dependency %s not found", key)
	}

	if write == nil {
		return errors.Errorf("dependency %s does not need or support secrets", key)
	}

	return errors.Wrapf(write(path[1:], value), "reading secret from dependency %s failed", key)
}

func ReadSecret(gitClient *git.Client, adapt AdaptFunc, path string) (string, error) {

	treeDesired, treeSecrets, err := parse(gitClient)
	if err != nil {
		return "", err
	}

	_, read, _, err := adapt(treeDesired, treeSecrets, &Tree{})
	if err != nil {
		return "", err
	}

	return read(strings.Split(path, "."))
}

func WriteSecret(gitClient *git.Client, adapt AdaptFunc, path, value string) error {

	treeDesired, treeSecrets, err := parse(gitClient)
	if err != nil {
		return err
	}

	_, _, write, err := adapt(treeDesired, treeSecrets, &Tree{})
	if err != nil {
		return err
	}

	if err := write(strings.Split(path, "."), value); err != nil {
		return err
	}

	return writeSecretsFunc(gitClient, treeSecrets)()
}

func parse(gitClient *git.Client) (desired *Tree, secrets *Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	rawDesired, err := gitClient.Read("desired.yml")
	if err != nil {
		return nil, nil, err
	}
	treeDesired := &Tree{}
	if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
		return nil, nil, err
	}

	rawSecrets, err := gitClient.Read("secrets.yml")
	if err != nil {
		return nil, nil, err
	}

	treeSecrets := &Tree{}
	if err := yaml.Unmarshal([]byte(rawSecrets), treeSecrets); err != nil {
		return nil, nil, err
	}

	return treeDesired, treeSecrets, nil
}

func writeSecretsFunc(gitClient *git.Client, secrets *Tree) PushSecretsFunc {
	return func() error {
		return gitClient.UpdateRemote(git.File{
			Path:    "secrets.yml",
			Content: common.MarshalYAML(secrets),
		})
	}
}
