package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(logger logging.Logger, masterkey string, id string) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		secretsKind := &SecretsV0{Common: secretsTree.Common}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		lbCurrent := &orbiter.Tree{}
		var lbEnsurer orbiter.EnsureFunc
		switch desiredKind.Deps.Common.Kind {
		//		case "orbiter.caos.ch/ExternalLoadBalancer":
		//			return []orbiter.Assembler{external.New(depPath, generalOverwriteSpec, externallbadapter.New())}, nil
		case "orbiter.caos.ch/DynamicLoadBalancer":
			if lbEnsurer, err = dynamic.AdaptFunc(desiredKind.Spec.RemoteUser)(desiredKind.Deps, nil, lbCurrent); err != nil {
				return nil, err
			}
			//		return []orbiter.Assembler{dynamic.New(depPath, generalOverwriteSpec, dynamiclbadapter.New(kind.Spec.RemoteUser))}, nil
		default:
			return nil, errors.Errorf("unknown loadbalancing kind %s", desiredKind.Deps.Common.Kind)
		}

		current := &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/StaticProvider",
				Version: "v0",
			},
			Deps: lbCurrent,
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent, nodeAgentsDesired map[string]*orbiter.NodeAgentSpec) (err error) {
			defer func() {
				err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
			}()
			if err := lbEnsurer(nodeAgentsCurrent, nodeAgentsDesired); err != nil {
				return err
			}
			return ensure(desiredKind, current, secretsKind, nodeAgentsDesired, lbCurrent, masterkey, logger, id)
		}, nil
	}
}
