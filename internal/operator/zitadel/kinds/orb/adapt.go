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

		iamCurrent := &tree.Tree{}
		queryIAM, destroyIAM, err := iam.AdaptFunc(timestamp, features...)(orbMonitor, desiredKind.IAM, iamCurrent)
		if err != nil {
			return nil, nil, err
		}

		queriers := []zitadel.QueryFunc{queryIAM}
		destroyers := []zitadel.DestroyFunc{destroyIAM}

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/Orb",
				Version: "v0",
			},
			IAM: iamCurrent,
		}

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				return zitadel.QueriersToEnsureFunc(monitor, true, queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(monitor, destroyers),
			nil
	}
}
