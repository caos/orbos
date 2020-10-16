package provided

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/secret"
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
		map[string]*secret.Secret,
		error,
	) {
		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ProvidedDatabase",
				Version: "v0",
			},
		}
		current.Parsed = currentDB

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
				currentDB.Current.URL = desiredKind.Spec.URL
				currentDB.Current.Port = desiredKind.Spec.Port

				return func(k8sClient *kubernetes.Client) error {
					return nil
				}, nil
			}, func(k8sClient *kubernetes.Client) error {
				return nil
			},
			map[string]*secret.Secret{},
			nil
	}
}
