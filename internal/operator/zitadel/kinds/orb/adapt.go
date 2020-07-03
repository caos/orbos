package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc() zitadel.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc zitadel.QueryFunc, destroyFunc zitadel.DestroyFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind
		currentTree = &tree.Tree{}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		queriers := make([]zitadel.QueryFunc, 0)
		destroyers := make([]zitadel.DestroyFunc, 0)

		iamCurrent := &tree.Tree{}
		queryIAM, destroyIAM, err := iam.AdaptFunc()(monitor, desiredKind.IAM, iamCurrent)
		if err != nil {
			return nil, nil, err
		}

		queriers = append(queriers, queryIAM)
		destroyers = append(destroyers, destroyIAM)

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/Orb",
				Version: "v0",
			},
			IAM: iamCurrent,
		}

		return func(k8sClient *kubernetes.Client) (zitadel.EnsureFunc, error) {
				ensurers := make([]zitadel.EnsureFunc, 0)
				for _, querier := range queriers {
					ensurer, err := querier(k8sClient)
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func(k8sClient *kubernetes.Client) error {
					for _, ensurer := range ensurers {
						if err := ensurer(k8sClient); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func(k8sClient *kubernetes.Client) error {
				for _, destroyer := range destroyers {
					if err := destroyer(k8sClient); err != nil {
						return err
					}
				}
				return nil
			},
			nil

	}
}
