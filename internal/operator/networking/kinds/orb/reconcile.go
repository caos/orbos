package orb

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"github.com/caos/orbos/pkg/treelabels"
)

func Reconcile(
	monitor mntr.Monitor,
	spec *Spec,
	gitops bool,
) core.EnsureFunc {
	return func(k8sClient kubernetes.ClientInt) (err error) {
		recMonitor := monitor.WithField("version", spec.Version)

		if spec.Version == "" {
			err := errors.New("No version set in networking.yml")
			monitor.Error(err)
			return err
		}

		imageRegistry := spec.CustomImageRegistry
		if imageRegistry == "" {
			imageRegistry = "ghcr.io"
		}

		if spec.SelfReconciling {
			desiredTree := &tree.Tree{
				Common: tree.NewCommon("networking.caos.ch/Orb", "v0", false),
			}

			if err := kubernetes.EnsureNetworkingArtifacts(monitor, treelabels.MustForAPI(desiredTree, mustDatabaseOperator(&spec.Version)), k8sClient, spec.Version, spec.NodeSelector, spec.Tolerations, imageRegistry, gitops); err != nil {
				return fmt.Errorf("failed to deploy networking-operator into k8s-cluster: %w", err)
			}

			recMonitor.Info("Applied networking-operator")
		}
		return nil

	}
}
