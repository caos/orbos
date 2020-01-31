package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(logger logging.Logger, masterkey string, id string) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, destroyFunc orbiter.DestroyFunc, secrets map[string]*orbiter.Secret, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}

		if secretsTree.Common == nil {
			secretsTree.Common = &orbiter.Common{
				Kind:    "orbiter.caos.ch/StaticProvider",
				Version: "v0",
			}
		}

		secretsKind := &SecretsV0{
			Common: secretsTree.Common,
			Secrets: Secrets{
				BootstrapKeyPrivate:   &orbiter.Secret{Masterkey: masterkey},
				BootstrapKeyPublic:    &orbiter.Secret{Masterkey: masterkey},
				MaintenanceKeyPrivate: &orbiter.Secret{Masterkey: masterkey},
				MaintenanceKeyPublic:  &orbiter.Secret{Masterkey: masterkey},
			},
		}
		if secretsTree.Original != nil {
			if err := secretsTree.Original.Decode(secretsKind); err != nil {
				return nil, nil, nil, migrate, errors.Wrap(err, "parsing secrets failed")
			}
		}
		if secretsKind.Secrets.BootstrapKeyPrivate == nil {
			secretsKind.Secrets.BootstrapKeyPrivate = &orbiter.Secret{Masterkey: masterkey}
		}

		if secretsKind.Secrets.BootstrapKeyPublic == nil {
			secretsKind.Secrets.BootstrapKeyPublic = &orbiter.Secret{Masterkey: masterkey}
		}

		if secretsKind.Secrets.MaintenanceKeyPrivate == nil {
			secretsKind.Secrets.MaintenanceKeyPrivate = &orbiter.Secret{Masterkey: masterkey}
		}

		if secretsKind.Secrets.MaintenanceKeyPublic == nil {
			secretsKind.Secrets.MaintenanceKeyPublic = &orbiter.Secret{Masterkey: masterkey}
		}

		secretsTree.Parsed = secretsKind

		lbCurrent := &orbiter.Tree{}
		switch desiredKind.Loadbalancing.Common.Kind {
		//		case "orbiter.caos.ch/ExternalLoadBalancer":
		//			return []orbiter.Assembler{external.New(depPath, generalOverwriteSpec, externallbadapter.New())}, nil
		case "orbiter.caos.ch/DynamicLoadBalancer":
			_, _, _, lMigrate, err := dynamic.AdaptFunc(logger)(desiredKind.Loadbalancing, nil, lbCurrent)
			if err != nil {
				return nil, nil, nil, migrate, err
			}
			if lMigrate {
				migrate = true
			}
			//		return []orbiter.Assembler{dynamic.New(depPath, generalOverwriteSpec, dynamiclbadapter.New(kind.Spec.RemoteUser))}, nil
		default:
			return nil, nil, nil, migrate, errors.Errorf("unknown loadbalancing kind %s", desiredKind.Loadbalancing.Common.Kind)
		}

		current := &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/StaticProvider",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error) {
				return errors.Wrapf(ensure(desiredKind, current, secretsKind, psf, nodeAgentsDesired, lbCurrent.Parsed, masterkey, logger, id), "ensuring %s failed", desiredKind.Common.Kind)
			}, func() error {
				return destroy(logger, desiredKind, current, secretsKind.Secrets, id)
			}, map[string]*orbiter.Secret{
				"bootstrapkeyprivate":   secretsKind.Secrets.BootstrapKeyPrivate,
				"bootstrapkeypublic":    secretsKind.Secrets.BootstrapKeyPublic,
				"maintenancekeyprivate": secretsKind.Secrets.MaintenanceKeyPrivate,
				"maintenancekeypublic":  secretsKind.Secrets.MaintenanceKeyPublic,
			}, migrate, nil
	}
}
