package provided

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc() zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {
		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		current.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ProvidedDatabase",
				Version: "v0",
			},
			Spec: desiredKind.Spec,
		}

		return func(k8sClient *kubernetes.Client) (zitadel.EnsureFunc, error) {
				return func(k8sClient *kubernetes.Client) error {
					return nil
				}, nil
			}, func(k8sClient *kubernetes.Client) error {
				return nil
			},
			nil
	}
}
