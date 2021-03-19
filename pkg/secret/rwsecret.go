package secret

import (
	"fmt"
	"sort"
	"strings"

	v1 "k8s.io/api/core/v1"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"

	macherrs "k8s.io/apimachinery/pkg/api/errors"

	"github.com/caos/orbos/pkg/kubernetes"

	"github.com/AlecAivazis/survey/v2"

	"github.com/caos/orbos/internal/api"

	"github.com/caos/orbos/pkg/tree"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
)

type PushFunc func(gitClient *git.Client, desired *tree.Tree) api.PushDesiredFunc
type PushFuncs func(trees map[string]*tree.Tree, path string) error
type GetFuncs func() (map[string]*Secret, map[string]*Existing, map[string]*tree.Tree, error)

func Read(
	k8sClient kubernetes.ClientInt,
	path string,
	getFunc GetFuncs,
) (
	val string,
	err error,
) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("reading secret failed: %w", err)
		}
	}()

	allSecrets, allExisting, _, err := getFunc()
	if err != nil {
		return "", err
	}
	if k8sClient == nil {
		allExisting = make(map[string]*Existing)
	}

	/*
		if allSecrets == nil || len(allSecrets) == 0 {
			return "", errors.New("no secrets found")
		}
	*/

	secret, err := findSecret(allSecrets, allExisting, &path, false)
	if err != nil {
		return "", err
	}

	switch secretType := secret.(type) {
	case *Secret:
		if secretType.Value == "" {
			return "", fmt.Errorf("secret %s is empty", path)
		}
		return secretType.Value, nil
	case *Existing:
		if secretType.Name == "" {
			return "", fmt.Errorf("secret %s has no name specified", path)
		}
		if secretType.Key == "" {
			return "", fmt.Errorf("secret %s has no key specified", path)
		}
		k8sSecret, err := k8sClient.GetSecret(existingSecretsNamespace, secretType.Name)
		if err != nil {
			return "", err
		}
		bytes, ok := k8sSecret.Data[secretType.Key]
		if !ok || len(bytes) == 0 {
			return "", fmt.Errorf("Kubernetes secret is empty at key %s", secretType.Key)
		}
		return string(bytes), nil
	}
	panic(fmt.Errorf("unknown secret of type %T", secret))
}

func Rewrite(
	monitor mntr.Monitor,
	gitClient *git.Client,
	newMasterKey string,
	desired *tree.Tree,
	pushFunc PushFunc,
) error {
	oldMasterKey := Masterkey
	Masterkey = newMasterKey
	defer func() {
		Masterkey = oldMasterKey
	}()

	return pushFunc(gitClient, desired)(monitor)
}

func Write(
	monitor mntr.Monitor,
	k8sClient kubernetes.ClientInt,
	path,
	value,
	writtenByCLI,
	writtenByVersion string,
	getFunc GetFuncs,
	pushFunc PushFuncs,
) error {
	allSecrets, allExisting, allTrees, err := getFunc()
	if err != nil {
		return err
	}

	if k8sClient == nil {
		allExisting = make(map[string]*Existing)
	}

	secret, err := findSecret(allSecrets, allExisting, &path, true)
	if err != nil {
		return err
	}

	switch secretType := secret.(type) {
	case *Secret:
		if secretType.Value == value {
			monitor.Info("Value is unchanged")
			return nil
		}
		secretType.Value = value
	case *Existing:
		var refChanged bool
		if secretType.Name == "" {
			secretType.Name = strings.ReplaceAll(path, ".", "-")
			refChanged = true
		}

		if secretType.Key == "" {
			secretType.Key = "default"
			refChanged = true
		}

		k8sSecret, err := k8sClient.GetSecret(existingSecretsNamespace, secretType.Name)
		if macherrs.IsNotFound(err) {
			err = nil
			k8sSecret = &v1.Secret{
				ObjectMeta: mach.ObjectMeta{
					Name:      secretType.Name,
					Namespace: existingSecretsNamespace,
					Labels: map[string]string{
						"cli":     writtenByCLI,
						"version": writtenByVersion,
					},
				},
				Immutable: boolPtr(false),
				Type:      v1.SecretTypeOpaque,
			}
		}
		if err != nil {
			return err
		}

		if k8sSecret.Data == nil {
			k8sSecret.Data = make(map[string][]byte)
		}
		k8sSecret.Data[secretType.Key] = []byte(value)
		if err := k8sClient.ApplySecret(k8sSecret); err != nil {
			return err
		}
		if !refChanged {
			return nil
		}
	}

	return pushFunc(allTrees, path)
}

func GetOperatorSecrets(
	monitor mntr.Monitor,
	gitops bool,
	allTrees map[string]*tree.Tree,
	allSecrets map[string]*Secret,
	allExistingSecrets map[string]*Existing,
	operator string,
	yamlExistsInGit func() (bool, error),
	treeFromGit func() (*tree.Tree, error),
	treeFromCRD func() (*tree.Tree, error),
	getOperatorSpecifics func(*tree.Tree) (map[string]*Secret, map[string]*Existing, bool, error),
) error {

	if gitops {
		foundGitYAML, err := yamlExistsInGit()
		if err != nil {
			return err
		}

		if !foundGitYAML {
			monitor.Info(fmt.Sprintf("no file for %s found", operator))
			return nil
		}

		operatorTree, err := treeFromGit()
		if err != nil {
			return err
		}
		allTrees[operator] = operatorTree
	} else {
		operatorTree, err := treeFromCRD()
		if operatorTree == nil {
			return err
		}
		allTrees[operator] = operatorTree
	}

	secrets, existing, migrate, err := getOperatorSpecifics(allTrees[operator])
	if err != nil {
		return err
	}

	if migrate {
		return fmt.Errorf("please use the api command to migrate to the latest %s api first", operator)
	}

	if !gitops {
		secrets = nil
	}

	suffixedSecrets := make(map[string]*Secret, len(secrets))
	suffixedExisting := make(map[string]*Existing, len(existing))
	for k, v := range secrets {
		suffixedSecrets[k+".encrypted"] = v
	}
	for k, v := range existing {
		suffixedExisting[k+".existing"] = v
	}

	AppendSecrets(operator, allSecrets, suffixedSecrets, allExistingSecrets, suffixedExisting)

	return nil
}

func secretsListToSlice(
	secrets map[string]*Secret,
	existing map[string]*Existing,
	includeEmpty bool,
) []string {
	items := make([]string, 0, len(secrets)+len(existing))
	for key, value := range secrets {
		if includeEmpty || (value != nil && value.Value != "") {
			items = append(items, key)
		}
	}
	for key, value := range existing {
		if includeEmpty || (value != nil && value.Name != "" && value.Key != "") {
			items = append(items, key)
		}
	}
	return items
}

func findSecret(
	allSecrets map[string]*Secret,
	allExisting map[string]*Existing,
	path *string,
	includeEmpty bool,
) (
	interface{},
	error,
) {
	if *path != "" {
		return exactSecret(allSecrets, allExisting, *path)
	}

	selectItems := secretsListToSlice(allSecrets, allExisting, includeEmpty)

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

	return exactSecret(allSecrets, allExisting, *path)
}

func exactSecret(
	secrets map[string]*Secret,
	existings map[string]*Existing,
	path string,
) (
	interface{},
	error,
) {
	secret, ok := secrets[path]
	if ok {
		return secret, nil
	}

	existing, ok := existings[path]
	if ok {
		return existing, nil
	}

	return nil, fmt.Errorf("no secret found at %s", path)
}

func boolPtr(b bool) *bool { return &b }
