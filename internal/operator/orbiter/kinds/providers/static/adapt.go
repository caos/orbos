package static

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(id string, whitelist dynamic.WhiteListFunc) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, finishedChan chan bool, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind, err := parseDesiredV0(desiredTree)
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

		lbCurrent := &tree.Tree{}
		var lbQuery orbiter.QueryFunc

		lbQuery, _, migrateLocal, err := loadbalancers.GetQueryAndDestroyFunc(monitor, whitelist, desiredKind.Loadbalancing, lbCurrent, finishedChan)
		if err != nil {
			return nil, nil, migrate, err
		}
		if migrateLocal {
			migrate = true
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

				lbQueryFunc := func() (orbiter.EnsureFunc, error) {
					return lbQuery(nodeAgentsCurrent, nodeAgentsDesired, nil)
				}

				if _, err := orbiter.QueryFuncGoroutine(lbQueryFunc); err != nil {
					return nil, err
				}

				queryFunc := func() (orbiter.EnsureFunc, error) {
					return query(desiredKind, current, nodeAgentsDesired, lbCurrent.Parsed, monitor, id)
				}
				return orbiter.QueryFuncGoroutine(queryFunc)
			}, func() error {
				return destroy(monitor, desiredKind, current, id)
			}, migrate, nil
	}
}
