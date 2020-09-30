package orb

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func Reconcile(monitor mntr.Monitor, desiredTree *tree.Tree, version string) core.EnsureFunc {
	return func(k8sClient *kubernetes.Client) (err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		recMonitor := monitor.WithField("version", desiredKind.Spec.Version)

		networkingVersion := version
		if desiredKind.Spec.Version != "" {
			networkingVersion = desiredKind.Spec.Version
		} else {
			monitor.Info(fmt.Sprintf("No version set in zitadel.yml, so default version %s will get applied", version))
		}

		if err := kubernetes.EnsureNetworkingArtifacts(monitor, k8sClient, networkingVersion, desiredKind.Spec.NodeSelector, desiredKind.Spec.Tolerations); err != nil {
			recMonitor.Error(errors.Wrap(err, "Failed to deploy zitadel-operator into k8s-cluster"))
			return err
		}

		recMonitor.Info("Applied zitadel-operator")
		return nil

	}
}
