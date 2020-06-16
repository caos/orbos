package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/push"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	orb *orb.Orb,
	orbiterCommit string,
	oneoff bool,
	deployOrbiter bool) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, finishedChan chan bool, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if err := desiredKind.validate(); err != nil {
			return nil, nil, migrate, err
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		providerCurrents := make(map[string]*tree.Tree)
		providerQueriers := make([]orbiter.QueryFunc, 0)
		providerDestroyers := make([]orbiter.DestroyFunc, 0)

		whitelistChan := make(chan []*orbiter.CIDR)

		for provID, providerTree := range desiredKind.Providers {

			providerCurrent := &tree.Tree{}
			providerCurrents[provID] = providerCurrent

			//			providermonitor := monitor.WithFields(map[string]interface{}{
			//				"provider": provID,
			//			})

			//			providerID := id + provID
			query, destroy, migrateLocal, err := providers.GetQueryAndDestroyFuncs(
				monitor,
				orb,
				provID,
				providerTree,
				providerCurrent,
				whitelistChan,
				finishedChan,
				orbiterCommit, orb.URL, orb.Repokey,
			)

			if err != nil {
				return nil, nil, migrate, err
			}

			if migrateLocal {
				migrate = true
			}

			providerQueriers = append(providerQueriers, query)
			providerDestroyers = append(providerDestroyers, destroy)
		}

		var provCurr map[string]interface{}
		destroyProviders := func() (map[string]interface{}, error) {
			if provCurr != nil {
				return provCurr, nil
			}

			provCurr = make(map[string]interface{})
			for _, destroyer := range providerDestroyers {
				if err := destroyer(); err != nil {
					return nil, err
				}
			}

			for currKey, currVal := range providerCurrents {
				provCurr[currKey] = currVal.Parsed
			}
			return provCurr, nil
		}

		clusterCurrents := make(map[string]*tree.Tree)
		clusterQueriers := make([]orbiter.QueryFunc, 0)
		clusterDestroyers := make([]orbiter.DestroyFunc, 0)
		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterCurrent := &tree.Tree{}
			clusterCurrents[clusterID] = clusterCurrent
			query, destroy, migrateLocal, err := clusters.GetQueryAndDestroyFuncs(
				monitor,
				orb,
				clusterID,
				clusterTree,
				orbiterCommit,
				oneoff,
				deployOrbiter,
				clusterCurrent,
				destroyProviders,
				whitelistChan,
				finishedChan,
			)

			if err != nil {
				return nil, nil, migrate, err
			}
			clusterQueriers = append(clusterQueriers, query)
			clusterDestroyers = append(clusterDestroyers, destroy)
			if migrateLocal {
				migrate = true
			}
		}

		currentTree.Parsed = &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/Orb",
				Version: "v0",
			},
			Clusters:  clusterCurrents,
			Providers: providerCurrents,
		}

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, _ map[string]interface{}) (ensureFunc orbiter.EnsureFunc, err error) {

				providerEnsurers := make([]orbiter.EnsureFunc, 0)
				queriedProviders := make(map[string]interface{})
				for _, querier := range providerQueriers {
					queryFunc := func() (orbiter.EnsureFunc, error) {
						return querier(nodeAgentsCurrent, nodeAgentsDesired, nil)
					}
					ensurer, err := orbiter.QueryFuncGoroutine(queryFunc)

					if err != nil {
						return nil, err
					}
					providerEnsurers = append(providerEnsurers, ensurer)
				}

				for currKey, currVal := range providerCurrents {
					queriedProviders[currKey] = currVal.Parsed
				}

				clusterEnsurers := make([]orbiter.EnsureFunc, 0)
				for _, querier := range clusterQueriers {
					queryFunc := func() (orbiter.EnsureFunc, error) {
						return querier(nodeAgentsCurrent, nodeAgentsDesired, queriedProviders)
					}
					ensurer, err := orbiter.QueryFuncGoroutine(queryFunc)

					if err != nil {
						return nil, err
					}
					clusterEnsurers = append(clusterEnsurers, ensurer)
				}

				return func(psf push.Func) *orbiter.EnsureResult {
					defer func() {
						err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
					}()

					for _, ensurer := range append(providerEnsurers, clusterEnsurers...) {
						ensureFunc := func() *orbiter.EnsureResult {
							return ensurer(psf)
						}

						if result := orbiter.EnsureFuncGoroutine(ensureFunc); result.Err != nil || !result.Done {
							return result
						}
					}

					return nil
				}, nil
			}, func() error {
				defer func() {
					err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
				}()

				for _, destroyer := range clusterDestroyers {
					if err := orbiter.DestroyFuncGoroutine(destroyer); err != nil {
						return err
					}
				}
				return nil
			}, migrate, nil
	}
}
