package orbiter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbiter/internal/git"
)

func JoinPath(base string, append ...string) string {
	for _, item := range append {
		base = fmt.Sprintf("%s.%s", base, item)
	}
	return base
}

func ReadSecret(gitClient *git.Client, adapt AdaptFunc, path string) (string, error) {

	secret, _, err := findSecret(gitClient, adapt, path, func(secrets map[string]*Secret) []string {
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

func WriteSecret(gitClient *git.Client, adapt AdaptFunc, path, value string) error {

	secret, tree, err := findSecret(gitClient, adapt, path, func(secrets map[string]*Secret) []string {
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

	return pushSecretsFunc(gitClient, tree)()
}

func findSecret(gitClient *git.Client, adapt AdaptFunc, path string, items func(map[string]*Secret) []string) (*Secret, *Tree, error) {
	treeDesired, err := parse(gitClient)
	if err != nil {
		return nil, nil, err
	}

	_, _, secrets, _, err := adapt(treeDesired, &Tree{})
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
