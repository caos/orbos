package static

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/common"
	"github.com/caos/orbos/v5/internal/operator/orbiter"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/v5/internal/ssh"
	"github.com/caos/orbos/v5/mntr"
	orbcfg "github.com/caos/orbos/v5/pkg/orb"
	"github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
)

func AdaptFunc(
	id string,
	whitelist dynamic.WhiteListFunc,
	orbiterCommit,
	repoURL,
	repoKey string,
	pprof bool,
) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, finishedChan chan struct{}, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, configureFunc orbiter.ConfigureFunc, migrate bool, secrets map[string]*secret.Secret, err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("building %s failed: %w", desiredTree.Common.Kind, err)
			}
		}()
		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, nil, migrate, nil, fmt.Errorf("parsing desired state failed: %w", err)
		}
		desiredTree.Parsed = desiredKind
		secrets = make(map[string]*secret.Secret, 0)
		secret.AppendSecrets("", secrets, getSecretsMap(desiredKind), nil, nil)

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		if desiredKind.Spec.ExternalInterfaces == nil {
			desiredKind.Spec.ExternalInterfaces = make([]string, 0)
			migrate = true
		}

		if desiredKind.Spec.PrivateInterface == "" {
			desiredKind.Spec.PrivateInterface = "eth0"
			migrate = true
		}

		if err := desiredKind.validateAdapt(); err != nil {
			return nil, nil, nil, migrate, nil, err
		}

		lbCurrent := &tree.Tree{}
		var lbQuery orbiter.QueryFunc

		lbQuery, lbDestroy, lbConfigure, migrateLocal, lbsecrets, err := loadbalancers.GetQueryAndDestroyFunc(monitor, whitelist, desiredKind.Loadbalancing, lbCurrent, finishedChan)
		if err != nil {
			return nil, nil, nil, migrate, nil, err
		}
		if migrateLocal {
			migrate = true
		}
		secret.AppendSecrets("", secrets, lbsecrets, nil, nil)

		current := &Current{
			Common: tree.NewCommon("orbiter.caos.ch/StaticProvider", "v0", false),
		}
		current.Current.privateInterface = desiredKind.Spec.PrivateInterface
		currentTree.Parsed = current

		svc := NewMachinesService(monitor, desiredKind, id)
		return func(nodeAgentsCurrent *common.CurrentNodeAgents, nodeAgentsDesired *common.DesiredNodeAgents, _ map[string]interface{}) (ensureFunc orbiter.EnsureFunc, err error) {
				defer func() {
					if err != nil {
						err = fmt.Errorf("querying %s failed: %w", desiredKind.Common.Kind, err)
					}
				}()

				if err := desiredKind.validateQuery(); err != nil {
					return nil, err
				}

				if _, err := lbQuery(nodeAgentsCurrent, nodeAgentsDesired, nil); err != nil {
					return nil, err
				}

				if err := svc.updateKeys(); err != nil {
					return nil, err
				}
				_, iterateNA := core.NodeAgentFuncs(monitor, repoURL, repoKey, pprof)
				return query(desiredKind, current, nodeAgentsDesired, nodeAgentsCurrent, lbCurrent.Parsed, monitor, svc, iterateNA, orbiterCommit)
			}, func(delegates map[string]interface{}) error {
				if err := lbDestroy(delegates); err != nil {
					return err
				}

				if err := svc.updateKeys(); err != nil {
					return err
				}

				return destroy(svc, desiredKind, current)
			}, func(orb orbcfg.Orb) error {
				if err := lbConfigure(orb); err != nil {
					return err
				}

				initKeys := desiredKind.Spec.Keys == nil
				if initKeys ||
					desiredKind.Spec.Keys.MaintenanceKeyPrivate == nil || desiredKind.Spec.Keys.MaintenanceKeyPrivate.Value == "" ||
					desiredKind.Spec.Keys.MaintenanceKeyPublic == nil || desiredKind.Spec.Keys.MaintenanceKeyPublic.Value == "" {
					if initKeys {
						desiredKind.Spec.Keys = &Keys{}
					}
					priv, pub := ssh.Generate()
					desiredKind.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{Value: priv}
					desiredKind.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{Value: pub}
					return nil
				}

				return core.ConfigureNodeAgents(svc, monitor, orb, pprof)
			},
			migrate,
			secrets,
			nil
	}
}
