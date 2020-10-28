package deployment

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func ReadyFunc(
	monitor mntr.Monitor,
	namespace string,
) zitadel.EnsureFunc {
	return func(k8sClient *kubernetes.Client) error {
		monitor.Info("waiting for deployment to be ready")
		if err := k8sClient.WaitUntilDeploymentReady(namespace, deployName, true, true, 60); err != nil {
			monitor.Error(errors.Wrap(err, "error while waiting for deployment to be ready"))
			return err
		}
		monitor.Info("deployment is ready")
		return nil
	}
}
