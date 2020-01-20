package orbiter

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/git"
)

type ReadSecretFunc func(path []string) (string, error)

type WriteSecretFunc func(path []string, value string) error

func AdaptReadSecret(path []string, deps map[string]ReadSecretFunc, mapping map[string]*Secret) (string, error) {

	if len(path) == 0 {
		return "", errors.New("no path provided")
	}

	key := path[0]

	if len(path) == 1 {
		if len(mapping) == 0 {
			return "", errors.New("kind does not need or support secrets")
		}

		secret, ok := mapping[key]
		if !ok {
			return "", errors.Errorf("unknown secret %s", key)
		}
		return secret.Value, nil
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

	if len(path) == 1 {
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

	_, _, read, _, err := adapt(treeDesired, treeSecrets, &Tree{})
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

	_, _, _, write, err := adapt(treeDesired, treeSecrets, &Tree{})
	if err != nil {
		return err
	}

	if err := write(strings.Split(path, "."), value); err != nil {
		return err
	}

	return pushSecretsFunc(gitClient, treeSecrets)()
}
