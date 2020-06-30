package zitadel

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, error)

type EnsureFunc func(k8sClient *kubernetes.Client) error

type DestroyFunc func(k8sClient *kubernetes.Client) error

type QueryFunc func(k8sClient *kubernetes.Client) (EnsureFunc, error)

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
