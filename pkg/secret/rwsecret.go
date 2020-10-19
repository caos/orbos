package secret

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/caos/orbos/internal/api"

	"github.com/caos/orbos/pkg/tree"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
)

type PushFunc func(gitClient *git.Client, desired *tree.Tree) api.PushDesiredFunc
type PushFuncs func(monitor mntr.Monitor, gitClient *git.Client, trees map[string]*tree.Tree, path string) error
type GetFuncs func(monitor mntr.Monitor, gitClient *git.Client) (map[string]*Secret, map[string]*tree.Tree, error)

func Read(monitor mntr.Monitor, gitClient *git.Client, path string, getFunc GetFuncs) (string, error) {
	allSecrets, _, err := getFunc(monitor, gitClient)

	secret, err := findSecret(allSecrets, &path, false)
	if err != nil {
		return "", err
	}

	if secret.Value == "" {
		return "", fmt.Errorf("Secret %s is empty", path)
	}

	return secret.Value, nil
}

func Rewrite(monitor mntr.Monitor, gitClient *git.Client, newMasterKey string, desired *tree.Tree, pushFunc PushFunc) error {
	oldMasterKey := Masterkey
	Masterkey = newMasterKey
	defer func() {
		Masterkey = oldMasterKey
	}()

	return pushFunc(gitClient, desired)(monitor)
}

func Write(monitor mntr.Monitor, gitClient *git.Client, path, value string, getFunc GetFuncs, pushFunc PushFuncs) error {
	allSecrets, allTrees, err := getFunc(monitor, gitClient)

	secret, err := findSecret(allSecrets, &path, true)
	if err != nil {
		return err
	}

	secret.Value = value

	return pushFunc(monitor, gitClient, allTrees, path)
}

func secretsListToSlice(secrets map[string]*Secret, includeEmpty bool) []string {
	items := make([]string, 0, len(secrets))
	for key, value := range secrets {
		if includeEmpty || (value != nil && value.Value != "") {
			items = append(items, key)
		}
	}
	return items
}

func findSecret(allSecrets map[string]*Secret, path *string, includeEmpty bool) (*Secret, error) {
	if *path != "" {
		return exactSecret(allSecrets, *path)
	}

	selectItems := secretsListToSlice(allSecrets, includeEmpty)

	sort.Slice(selectItems, func(i, j int) bool {
		iDots := strings.Count(selectItems[i], ".")
		jDots := strings.Count(selectItems[j], ".")
		return iDots < jDots || iDots == jDots && selectItems[i] < selectItems[j]
	})

	var result string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a secret:",
		Options: selectItems,
	}, &result, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	*path = result

	return exactSecret(allSecrets, *path)
}

func exactSecret(secrets map[string]*Secret, path string) (*Secret, error) {
	secret, ok := secrets[path]
	if !ok {
		return nil, fmt.Errorf("Secret %s not found", path)
	}
	return secret, nil
}
