package secret

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/caos/orbos/internal/api"

	"github.com/caos/orbos/internal/tree"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/mntr"
)

const (
	boom    string = "boom"
	orbiter string = "orbiter"
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

type PushFunc func(gitClient *git.Client, desired *tree.Tree) api.PushDesiredFunc
type GetFunc func(monitor mntr.Monitor, gitClient *git.Client) (map[string]*Secret, map[string]*tree.Tree, error)
type Func func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*Secret, err error)

func JoinPath(base string, append ...string) string {
	for _, item := range append {
		base = fmt.Sprintf("%s.%s", base, item)
	}
	return base
}

func Read(monitor mntr.Monitor, gitClient *git.Client, path string, getFunc GetFunc) (string, error) {
	allSecrets, _, err := getFunc(monitor, gitClient)

	secret, err := findSecret(allSecrets, &path)
	if err != nil {
		return "", err
	}

	if secret.Value == "" {
		return "", fmt.Errorf("Secret %s not found", path)
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

func Write(monitor mntr.Monitor, gitClient *git.Client, path, value string, getFunc GetFunc) error {
	allSecrets, allTrees, err := getFunc(monitor, gitClient)

	secret, err := findSecret(allSecrets, &path)
	if err != nil {
		return err
	}

	secret.Value = value

	return pushTreeFromPrefix(monitor, gitClient, allTrees, path)
}

func secretsListToSlice(secrets map[string]*Secret) []string {
	items := make([]string, 0, len(secrets))
	for key := range secrets {
		items = append(items, key)
	}
	return items
}

func findSecret(allSecrets map[string]*Secret, path *string) (*Secret, error) {
	if *path != "" {
		return exactSecret(allSecrets, *path)
	}

	selectItems := secretsListToSlice(allSecrets)

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

func pushTreeFromPrefix(monitor mntr.Monitor, gitClient *git.Client, trees map[string]*tree.Tree, path string) error {
	operator := ""
	if strings.HasPrefix(path, orbiter) {
		operator = orbiter
	} else if strings.HasPrefix(path, boom) {
		operator = boom
	} else {
		return errors.New("Operator unknown")
	}

	desired, err := getTreeFromPrefix(trees, path)
	if err != nil {
		return err
	}

	if operator == "orbiter" {
		return api.PushOrbiterDesiredFunc(gitClient, desired)(monitor)
	} else if operator == "boom" {
		return api.PushBoomDesiredFunc(gitClient, desired)(monitor)
	}

	return errors.New("Operator unknown")
}

func getTreeFromPrefix(trees map[string]*tree.Tree, path string) (*tree.Tree, error) {
	operator := ""
	if strings.HasPrefix(path, orbiter) {
		operator = orbiter
	} else if strings.HasPrefix(path, boom) {
		operator = boom
	} else {
		return nil, errors.New("Operator unknown")
	}

	desired, found := trees[operator]
	if !found {
		return nil, errors.New("Operator unknown")
	}

	return desired, nil
}

func exactSecret(secrets map[string]*Secret, path string) (*Secret, error) {
	secret, ok := secrets[path]
	if !ok {
		return nil, fmt.Errorf("Secret %s not found", path)
	}
	return secret, nil
}
