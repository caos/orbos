package static

import (
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/mntr"
)

func AdaptFunc(masterkey string, id string) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
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
			return nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		if err := desiredKind.validate(); err != nil {
			return nil, nil, migrate, err
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

		lbCurrent := &tree.Tree{}
		var lbQuery orbiter.QueryFunc
		switch desiredKind.Loadbalancing.Common.Kind {
		//		case "orbiter.caos.ch/ExternalLoadBalancer":
		//			return []orbiter.Assembler{external.New(depPath, generalOverwriteSpec, externallbadapter.New())}, nil
		case "orbiter.caos.ch/DynamicLoadBalancer":
			lbQuery, _, migrate, err = dynamic.AdaptFunc()(monitor, desiredKind.Loadbalancing, lbCurrent)
			if err != nil {
				return nil, nil, migrate, err
			}
		default:
			return nil, nil, migrate, errors.Errorf("unknown loadbalancing kind %s", desiredKind.Loadbalancing.Common.Kind)
		}

		current := &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/StaticProvider",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, _ map[string]interface{}) (ensureFunc orbiter.EnsureFunc, err error) {
				defer func() {
					err = errors.Wrapf(err, "querying %s failed", desiredKind.Common.Kind)
				}()

				if _, err := lbQuery(nodeAgentsCurrent, nodeAgentsDesired, nil); err != nil {
					return nil, err
				}
				return query(desiredKind, current, nodeAgentsDesired, lbCurrent.Parsed, masterkey, monitor, id)
			}, func() error {
				return destroy(monitor, desiredKind, current, id)
			}, migrate, nil
	}
}
