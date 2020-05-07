package secret

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/caos/orbiter/internal/push"
	"github.com/caos/orbiter/internal/tree"
	"gopkg.in/yaml.v3"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/mntr"
)

const (
	boom    string = "boom"
	orbiter string = "orbiter"
	yml     string = "yml"
)

func Parse(gitClient *git.Client, files ...string) (trees []*tree.Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	for _, file := range files {
		raw, err := gitClient.Read(file)
		if err != nil {
			return nil, err
		}

		tree := &tree.Tree{}
		if err := yaml.Unmarshal([]byte(raw), tree); err != nil {
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
			if sec.Value != "" {
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

func Write(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, path, value string) error {

	secret, tree, operator, err := findSecret(monitor, gitClient, secretFunc, path, func(secrets map[string]*Secret) []string {
		items := make([]string, 0, len(secrets))
		for key := range secrets {
			items = append(items, key)
		}
		return items
	})
	if err != nil {
		return err
	}

	secret.Value = value

	return push.RewriteDesiredFunc(gitClient, tree, strings.Join([]string{operator, "yml"}, "."))(monitor)
}

func addSecretsPrefix(prefix string, secrets map[string]*Secret) map[string]*Secret {
	ret := make(map[string]*Secret, len(secrets))
	for k, v := range secrets {
		key := strings.Join([]string{prefix, k}, ".")
		ret[key] = v
	}

	return ret
}

func findSecret(monitor mntr.Monitor, gitClient *git.Client, secretFunc GetFunc, path string, items func(map[string]*Secret) []string) (*Secret, *tree.Tree, string, error) {
	getOperatorSecrets := func(operator string) (map[string]*Secret, *tree.Tree, error) {
		file := strings.Join([]string{operator, yml}, ".")
		trees, err := Parse(gitClient, file)
		if err != nil {
			return nil, nil, err
		}

		treeDesired := trees[0]
		secrets, err := secretFunc(operator)(monitor, treeDesired)
		if err != nil {
			return nil, nil, err
		}
		return addSecretsPrefix(operator, secrets), treeDesired, nil
	}

	secretsOrbiter, treeDesiredOrbiter, err := getOperatorSecrets(orbiter)
	if err != nil {
		return nil, nil, "", err
	}
	secretsAll := secretsOrbiter

	secretsBoom, treeDesiredBoom, err := getOperatorSecrets(boom)
	if err != nil {
		return nil, nil, "", err
	}
	for k, v := range secretsBoom {
		secretsAll[k] = v
	}

	if path != "" {
		operator := ""
		if strings.HasPrefix(path, orbiter) {
			operator = orbiter
		} else if strings.HasPrefix(path, boom) {
			operator = boom
		} else {
			return nil, nil, "", errors.New("Operator unknown")
		}
		secrets, treeDesired, err := getOperatorSecrets(operator)
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
		return sec, treeDesiredBoom, orbiter, err
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
