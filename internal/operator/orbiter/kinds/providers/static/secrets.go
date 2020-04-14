package static

import (
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/mntr"
)

func SecretsFunc(masterkey string) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind := &DesiredV0{
			Common: desiredTree.Common,
			Spec: Spec{
				Keys: Keys{
					BootstrapKeyPrivate:   &secret.Secret{Masterkey: masterkey},
					BootstrapKeyPublic:    &secret.Secret{Masterkey: masterkey},
					MaintenanceKeyPrivate: &secret.Secret{Masterkey: masterkey},
					MaintenanceKeyPublic:  &secret.Secret{Masterkey: masterkey},
				},
			},
		}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		if err := desiredKind.validate(); err != nil {
			return nil, err
		}

		if desiredKind.Spec.Keys.BootstrapKeyPrivate == nil {
			desiredKind.Spec.Keys.BootstrapKeyPrivate = &secret.Secret{Masterkey: masterkey}
		}

		if desiredKind.Spec.Keys.BootstrapKeyPublic == nil {
			desiredKind.Spec.Keys.BootstrapKeyPublic = &secret.Secret{Masterkey: masterkey}
		}

		if desiredKind.Spec.Keys.MaintenanceKeyPrivate == nil {
			desiredKind.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{Masterkey: masterkey}
		}

		if desiredKind.Spec.Keys.MaintenanceKeyPublic == nil {
			desiredKind.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{Masterkey: masterkey}
		}

		current := &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/StaticProvider",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return map[string]*secret.Secret{
			"bootstrapkeyprivate":   desiredKind.Spec.Keys.BootstrapKeyPrivate,
			"bootstrapkeypublic":    desiredKind.Spec.Keys.BootstrapKeyPublic,
			"maintenancekeyprivate": desiredKind.Spec.Keys.MaintenanceKeyPrivate,
			"maintenancekeypublic":  desiredKind.Spec.Keys.MaintenanceKeyPublic,
		}, nil
	}
}
