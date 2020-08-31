package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc(timestamp string, features ...string) zitadel.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc zitadel.QueryFunc, destroyFunc zitadel.DestroyFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		orbMonitor := monitor.WithField("kind", "orb")

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind
		currentTree = &tree.Tree{}

		if desiredKind.Spec.Verbose && !orbMonitor.IsVerbose() {
			orbMonitor = orbMonitor.Verbose()
		}

		query := zitadel.EnsureFuncToQueryFunc(func(k8sClient *kubernetes.Client) error {
			if err := kubernetes.EnsureZitadelArtifacts(monitor, k8sClient, desiredKind.Spec.Version, desiredKind.Spec.NodeSelector, desiredKind.Spec.Tolerations); err != nil {
				monitor.Error(errors.Wrap(err, "Failed to deploy zitadel-operator into k8s-cluster"))
				return err
			}
			return nil
		})

		iamCurrent := &tree.Tree{}
		queryIAM, destroyIAM, err := iam.GetQueryAndDestroyFuncs(orbMonitor, desiredKind.IAM, iamCurrent, timestamp, features...)
		if err != nil {
			return nil, nil, err
		}

		queriers := []zitadel.QueryFunc{
			query,
			queryIAM,
		}

		destroyers := []zitadel.DestroyFunc{
			destroyIAM,
		}

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/Orb",
				Version: "v0",
			},
			IAM: iamCurrent,
		}

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
				queried := map[string]interface{}{}
				monitor.WithField("queriers", len(queriers)).Info("Querying")
				return zitadel.QueriersToEnsureFunc(monitor, true, queriers, k8sClient, queried)
			},
			func(k8sClient *kubernetes.Client) error {
				monitor.WithField("destroyers", len(queriers)).Info("Destroy")
				return zitadel.DestroyersToDestroyFunc(monitor, destroyers)(k8sClient)
			},
			nil
	}
}
