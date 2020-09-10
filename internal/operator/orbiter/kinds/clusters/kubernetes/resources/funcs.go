package resources

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

type AdaptFuncToEnsure func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, error)
type AdaptFuncToDelete func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (DestroyFunc, error)

type EnsureFunc func(*kubernetes.Client) error

type DestroyFunc func(*kubernetes.Client) error

type QueryFunc func(*kubernetes.Client) (EnsureFunc, error)

func WrapFuncs(monitor mntr.Monitor, query QueryFunc, destroy DestroyFunc) (QueryFunc, DestroyFunc, error) {
	return func(client *kubernetes.Client) (ensureFunc EnsureFunc, err error) {
			monitor.Info("querying...")
			ensure, err := query(client)
			if err != nil {
				err := errors.Wrapf(err, "error while querying")
				monitor.Error(err)
				return nil, err
			}
			monitor.Info("queried")
			return func(k8sClient *kubernetes.Client) error {
				monitor.Info("ensuring...")
				if err := ensure(k8sClient); err != nil {
					return errors.Wrap(err, "error while destroying")
				}
				monitor.Info("ensured")
				return nil
			}, nil
		}, func(client *kubernetes.Client) error {
			monitor.Info("destroying...")
			err := destroy(client)
			if err != nil {
				err := errors.Wrapf(err, "error while destroying")
				monitor.Error(err)
				return err
			}
			monitor.Info("destroyed")
			return nil
		}, nil
}
