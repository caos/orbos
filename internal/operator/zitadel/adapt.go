package zitadel

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, error)

type EnsureFunc func(k8sClient *kubernetes.Client) error

type DestroyFunc func(k8sClient *kubernetes.Client) error

type QueryFunc func(k8sClient *kubernetes.Client, queried map[string]interface{}) (EnsureFunc, error)

func Parse(gitClient *git.Client, file string) (*tree.Tree, error) {
	if err := gitClient.Clone(); err != nil {
		return nil, err
	}

	tree := &tree.Tree{}
	if err := yaml.Unmarshal(gitClient.Read(file), tree); err != nil {
		return nil, err
	}

	return tree, nil
}

func ResourceDestroyToZitadelDestroy(destroyFunc resources.DestroyFunc) DestroyFunc {
	return func(k8sClient *kubernetes.Client) error {
		return destroyFunc(k8sClient)
	}
}

func ResourceQueryToZitadelQuery(queryFunc resources.QueryFunc) QueryFunc {
	return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (EnsureFunc, error) {
		ensure, err := queryFunc()
		ensureInternal := ResourceEnsureToZitadelEnsure(ensure)

		return func(k8sClient *kubernetes.Client) error {
			return ensureInternal(k8sClient)
		}, err
	}
}

func ResourceEnsureToZitadelEnsure(ensureFunc resources.EnsureFunc) EnsureFunc {
	return func(k8sClient *kubernetes.Client) error {
		return ensureFunc(k8sClient)
	}
}

func QueriersToEnsureFunc(queriers []QueryFunc, k8sClient *kubernetes.Client, queried map[string]interface{}) (EnsureFunc, error) {
	ensurers := make([]EnsureFunc, 0)
	for _, querier := range queriers {
		ensurer, err := querier(k8sClient, queried)
		if err != nil {
			return nil, err
		}
		ensurers = append(ensurers, ensurer)
	}

	return func(k8sClient *kubernetes.Client) error {
		for _, ensurer := range ensurers {
			if err := ensurer(k8sClient); err != nil {
				return err
			}
		}
		return nil
	}, nil
}
func DestroyersToDestroyFunc(destroyers []DestroyFunc) DestroyFunc {
	return func(k8sClient *kubernetes.Client) error {
		for _, destroyer := range destroyers {
			if err := destroyer(k8sClient); err != nil {
				return err
			}
		}
		return nil
	}
}
