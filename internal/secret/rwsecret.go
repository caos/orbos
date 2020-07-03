package secret

import (
	"errors"
	"fmt"
	"github.com/caos/orbos/internal/api"
	"sort"
	"strings"

	"github.com/caos/orbos/internal/tree"
	"gopkg.in/yaml.v3"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/mntr"
)

const (
	boom    string = "boom"
	orbiter string = "orbiter"
	zitadel string = "zitadel"
	yml     string = "yml"
)

func Parse(gitClient *git.Client, files ...string) (trees []*tree.Tree, err error) {
	for _, file := range files {
		tree := &tree.Tree{}
		if err := yaml.Unmarshal(gitClient.Read(file), tree); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	return trees, nil
}

type GetFunc func(operator string) Func
type Func func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*Secret, err error)

func JoinPath(base string, append ...string) string {
	for _, item := range append {
		base = fmt.Sprintf("%s.%s", base, item)
	}
	return base
}

func Read(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, path string) (string, error) {
	secret, _, _, err := findSecret(monitor, gitClient, secretFunc, path, func(secrets map[string]*Secret) []string {
		items := make([]string, 0)
		for key, sec := range secrets {
			if sec != nil && sec.Value != "" {
				items = append(items, key)
			}
		}
		return items
	})
	if err != nil {
		return "", err
	}

	if secret.Value == "" {
		return "", fmt.Errorf("Secret %s not found", path)
	}

	return secret.Value, nil
}

func Rewrite(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, operator string) error {
	tree, err := getTree(monitor, gitClient, secretFunc, operator)
	if err != nil {
		return err
	}
	if tree == nil {
		return nil
	}

	if operator == orbiter {
		return api.OrbiterSecretFunc(gitClient, tree)(monitor)
	} else if operator == boom {
		return api.BoomSecretFunc(gitClient, tree)(monitor)
	} else if operator == zitadel {
		return api.ZitadelSecretFunc(gitClient, tree)(monitor)
	}

	monitor.Info("No secrets written")
	return nil
}

func Write(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, path, value string) error {
	secret, tree, operator, err := findSecret(monitor, gitClient, secretFunc, path, secretsListToSlice)
	if err != nil {
		return err
	}

	if secret == nil {
		secret = &Secret{
			Value: value,
		}
	} else {
		secret.Value = value
	}

	if operator == orbiter {
		return api.OrbiterSecretFunc(gitClient, tree)(monitor)
	} else if operator == boom {
		return api.BoomSecretFunc(gitClient, tree)(monitor)
	} else if operator == zitadel {
		return api.ZitadelSecretFunc(gitClient, tree)(monitor)
	}

	monitor.Info("No secrets written")
	return nil
}

func addSecretsPrefix(prefix string, secrets map[string]*Secret) map[string]*Secret {
	ret := make(map[string]*Secret, len(secrets))
	if secrets != nil {
		for k, v := range secrets {
			key := strings.Join([]string{prefix, k}, ".")
			ret[key] = v
		}
	}

	return ret
}

func existsFileInGit(g *git.Client, path string) bool {
	of := g.Read(path)
	if of != nil && len(of) > 0 {
		return true
	}
	return false
}

func secretsListToSlice(secrets map[string]*Secret) []string {
	items := make([]string, 0, len(secrets))
	for key := range secrets {
		items = append(items, key)
	}
	return items
}

func getTree(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, operator string) (*tree.Tree, error) {
	_, treeDesired, err := getOperatorSecrets(monitor, operator, gitClient, secretFunc)

	return treeDesired, err
}

func getOperatorSecrets(monitor mntr.Monitor, operator string, gitClient *git.Client, secretFunc GetFunc) (map[string]*Secret, *tree.Tree, error) {
	file := strings.Join([]string{operator, yml}, ".")

	if existsFileInGit(gitClient, file) {
		trees, err := Parse(gitClient, file)
		if err != nil {
			return nil, nil, err
		}

		treeDesired := trees[0]
		secretsFunc := secretFunc(operator)
		if secretsFunc == nil {
			return nil, nil, errors.New("Operator unknown")
		}
		secrets, err := secretsFunc(monitor, treeDesired)
		if err != nil {
			return nil, nil, err
		}
		return addSecretsPrefix(operator, secrets), treeDesired, nil
	}
	return nil, nil, nil
}

func findSecret(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, path string, items func(map[string]*Secret) []string) (*Secret, *tree.Tree, string, error) {
	secretsAll := make(map[string]*Secret, 0)

	secretsOrbiter, treeDesiredOrbiter, err := getOperatorSecrets(monitor, orbiter, gitClient, secretFunc)
	if err != nil {
		return nil, nil, "", err
	}
	if secretsOrbiter != nil && len(secretsOrbiter) > 0 {
		for k, v := range secretsOrbiter {
			secretsAll[k] = v
		}
	}

	secretsBoom, treeDesiredBoom, err := getOperatorSecrets(monitor, boom, gitClient, secretFunc)
	if err != nil {
		return nil, nil, "", err
	}
	if secretsBoom != nil && len(secretsBoom) > 0 {
		for k, v := range secretsBoom {
			if k != "" && v != nil {
				secretsAll[k] = v
			}
		}
	}

	secretsZitadel, treeDesiredZitadel, err := getOperatorSecrets(monitor, zitadel, gitClient, secretFunc)
	if err != nil {
		return nil, nil, "", err
	}
	if secretsZitadel != nil && len(secretsZitadel) > 0 {
		for k, v := range secretsZitadel {
			if k != "" && v != nil {
				secretsAll[k] = v
			}
		}
	}

	if path != "" {
		operator := ""
		if strings.HasPrefix(path, orbiter) {
			operator = orbiter
		} else if strings.HasPrefix(path, boom) {
			operator = boom
		} else if strings.HasPrefix(path, zitadel) {
			operator = zitadel
		} else {
			return nil, nil, "", errors.New("Operator unknown")
		}
		secrets, treeDesired, err := getOperatorSecrets(monitor, operator, gitClient, secretFunc)
		sec, err := exactSecret(secrets, path)
		return sec, treeDesired, operator, err
	}

	selectItems := items(secretsAll)

	sort.Slice(selectItems, func(i, j int) bool {
		iDots := strings.Count(selectItems[i], ".")
		jDots := strings.Count(selectItems[j], ".")
		return iDots < jDots || iDots == jDots && selectItems[i] < selectItems[j]
	})

	prompt := promptui.Select{
		Label: "Select Secret",
		Items: selectItems,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return nil, nil, "", err
	}

	sec, err := exactSecret(secretsAll, result)
	if strings.HasPrefix(result, orbiter) {
		return sec, treeDesiredOrbiter, orbiter, err
	} else if strings.HasPrefix(result, boom) {
		return sec, treeDesiredBoom, boom, err
	} else if strings.HasPrefix(result, zitadel) {
		return sec, treeDesiredZitadel, zitadel, err
	}

	return nil, nil, "", errors.New("Operator unknown")
}

func exactSecret(secrets map[string]*Secret, path string) (*Secret, error) {
	secret, ok := secrets[path]
	if !ok {
		return nil, fmt.Errorf("Secret %s not found", path)
	}
	return secret, nil
}
