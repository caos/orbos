package secret

import (
	"fmt"
	"github.com/caos/orbiter/internal/push"
	"github.com/caos/orbiter/internal/tree"
	"gopkg.in/yaml.v3"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/mntr"
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

type Func func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*Secret, err error)

func JoinPath(base string, append ...string) string {
	for _, item := range append {
		base = fmt.Sprintf("%s.%s", base, item)
	}
	return base
}

func Read(monitor mntr.Monitor, gitClient *git.Client, secretFunc Func, path string) (string, error) {

	secret, _, err := findSecret(monitor, gitClient, secretFunc, path, func(secrets map[string]*Secret) []string {
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

func Write(monitor mntr.Monitor, gitClient *git.Client, secretFunc Func, path, value string) error {

	secret, tree, err := findSecret(monitor, gitClient, secretFunc, path, func(secrets map[string]*Secret) []string {
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

	return push.SecretsFunc(gitClient, tree)(monitor)
}

func findSecret(monitor mntr.Monitor, gitClient *git.Client, secretFunc Func, path string, items func(map[string]*Secret) []string) (*Secret, *tree.Tree, error) {
	trees, err := Parse(gitClient, "orbiter.yml")
	if err != nil {
		return nil, nil, err
	}

	treeDesired := trees[0]

	secrets, err := secretFunc(monitor, treeDesired)
	if err != nil {
		return nil, nil, err
	}

	if path != "" {
		sec, err := exactSecret(secrets, path)
		return sec, treeDesired, err
	}

	selectItems := items(secrets)

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
		return nil, nil, err
	}

	sec, err := exactSecret(secrets, result)
	return sec, treeDesired, err
}

func exactSecret(secrets map[string]*Secret, path string) (*Secret, error) {
	secret, ok := secrets[path]
	if !ok {
		return nil, fmt.Errorf("Secret %s not found", path)
	}
	return secret, nil
}
