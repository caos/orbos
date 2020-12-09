package orb

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func Reconcile(monitor mntr.Monitor, apiLabels *labels.API, desiredTree *tree.Tree) core.EnsureFunc {
	return func(k8sClient kubernetes.ClientInt) (err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		recMonitor := monitor.WithField("version", desiredKind.Spec.Version)

		if desiredKind.Spec.Version == "" {
			err := errors.New("No version set in database.yml")
			monitor.Error(err)
			return err
		}

		imageRegistry := desiredKind.Spec.CustomImageRegistry
		if imageRegistry == "" {
			imageRegistry = "ghcr.io"
		}

		if err := kubernetes.EnsureDatabaseArtifacts(monitor, apiLabels, k8sClient, desiredKind.Spec.Version, desiredKind.Spec.NodeSelector, desiredKind.Spec.Tolerations, imageRegistry); err != nil {
			recMonitor.Error(errors.Wrap(err, "Failed to deploy database-operator into k8s-cluster"))
			return err
		}

		recMonitor.Info("Applied database-operator")

		return nil

	}
}
