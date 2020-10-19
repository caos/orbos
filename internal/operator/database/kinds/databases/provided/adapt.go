package provided

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func AdaptFunc() func(
	monitor mntr.Monitor,
	desired *tree.Tree,
	current *tree.Tree,
) (
	core.QueryFunc,
	core.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		core.QueryFunc,
		core.DestroyFunc,
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

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (core.EnsureFunc, error) {
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
