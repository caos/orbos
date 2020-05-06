package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/mntr"
)

func AdaptFunc(masterkey, providerID, orbID string, whitelist dynamic.WhiteListFunc) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind, err := parseDesiredV0(desiredTree, masterkey)
		if err != nil {
			return nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		if err := desiredKind.validate(); err != nil {
			return nil, nil, migrate, err
		}

		initializeNecessarySecrets(desiredKind, masterkey)

		lbCurrent := &tree.Tree{}
		var lbQuery orbiter.QueryFunc

		lbQuery, _, migrateLocal, err := loadbalancers.GetQueryAndDestroyFunc(monitor, whitelist, desiredKind.Loadbalancing, lbCurrent)
		if err != nil {
			return nil, nil, migrate, err
		}
		if migrateLocal {
			migrate = true
		}

		current := &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/GCEProvider",
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
				return query(&desiredKind.Spec, current, monitor, nodeAgentsDesired, lbCurrent.Parsed, NewMachinesService(monitor, &desiredKind.Spec, providerID, orbID))
			}, func() error {
				return destroy(NewMachinesService(monitor, &desiredKind.Spec, providerID, orbID))
			}, migrate, nil
	}
}
