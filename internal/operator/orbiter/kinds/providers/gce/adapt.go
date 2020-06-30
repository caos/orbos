package gce

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(providerID, orbID string, whitelist dynamic.WhiteListFunc, orbiterCommit, repoURL, repoKey string, oneoff bool) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, finishedChan chan struct{}, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
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
				Kind:    "orbiter.caos.ch/GCEProvider",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent *common.CurrentNodeAgents, nodeAgentsDesired *common.DesiredNodeAgents, _ map[string]interface{}) (ensureFunc orbiter.EnsureFunc, err error) {
				defer func() {
					err = errors.Wrapf(err, "querying %s failed", desiredKind.Common.Kind)
				}()

				if _, err := lbQuery(nodeAgentsCurrent, nodeAgentsDesired, nil); err != nil {
					return nil, err
				}

				ctx, err := buildContext(monitor, &desiredKind.Spec, orbID, providerID, oneoff)
				if err != nil {
					return nil, err
				}
				return query(&desiredKind.Spec, current, lbCurrent.Parsed, ctx, nodeAgentsCurrent, nodeAgentsDesired, orbiterCommit, repoURL, repoKey)
			}, func() error {
				ctx, err := buildContext(monitor, &desiredKind.Spec, orbID, providerID, oneoff)
				if err != nil {
					return err
				}

				return destroy(ctx)
			}, migrate, nil
	}
}
