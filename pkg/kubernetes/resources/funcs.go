package resources

import (
	"fmt"

	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/tree"
)

type AdaptFuncToEnsure func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, error)
type AdaptFuncToDelete func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (DestroyFunc, error)

type EnsureFunc func(kubernetes.ClientInt) error

type DestroyFunc func(kubernetes.ClientInt) error

type QueryFunc func(kubernetes.ClientInt) (EnsureFunc, error)

func WrapFuncs(monitor mntr.Monitor, query QueryFunc, destroy DestroyFunc) (QueryFunc, DestroyFunc, error) {
	return func(client kubernetes.ClientInt) (ensureFunc EnsureFunc, err error) {
			monitor.Info("querying...")
			ensure, err := query(client)
			if err != nil {
				return nil, fmt.Errorf("error while querying: %w", err)
			}
			monitor.Info("queried")
			return func(k8sClient kubernetes.ClientInt) error {
				monitor.Info("ensuring...")
				if err := ensure(k8sClient); err != nil {
					return fmt.Errorf("error while destroying: %w", err)
				}
				monitor.Info("ensured")
				return nil
			}, nil
		}, func(client kubernetes.ClientInt) error {
			monitor.Info("destroying...")
			err := destroy(client)
			if err != nil {
				return fmt.Errorf("error while destroying: %w", err)
			}
			monitor.Info("destroyed")
			return nil
		}, nil
}
