package cmd

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func Reconcile(monitor mntr.Monitor, desiredTree *tree.Tree, version string) zitadel.EnsureFunc {
	return func(k8sClient *kubernetes.Client) (err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := orb.ParseDesiredV0(desiredTree)
		if err != nil {
			return errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		recMonitor := monitor.WithField("version", desiredKind.Spec.Version)

		zitadelVersion := version
		if desiredKind.Spec.Version != "" {
			zitadelVersion = desiredKind.Spec.Version
		} else {
			monitor.Info(fmt.Sprintf("No version set in zitadel.yml, so default version %s will get applied", version))
		}

		if err := kubernetes.EnsureZitadelArtifacts(monitor, k8sClient, zitadelVersion, desiredKind.Spec.NodeSelector, desiredKind.Spec.Tolerations); err != nil {
			recMonitor.Error(errors.Wrap(err, "Failed to deploy zitadel-operator into k8s-cluster"))
			return err
		}

		recMonitor.Info("Applied zitadel-operator")
		return nil

	}
}
