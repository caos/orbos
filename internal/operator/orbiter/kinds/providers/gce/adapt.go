package gce

import (
	"encoding/json"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"
	"google.golang.org/api/option"

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

				machinesSvc, addressesSvc, err := services(monitor, &desiredKind.Spec, orbID, providerID)
				if err != nil {
					return nil, err
				}

				return query(
					&desiredKind.Spec,
					current,
					monitor,
					nodeAgentsDesired,
					lbCurrent.Parsed,
					machinesSvc,
					addressesSvc,
				)
			}, func() error {
				machinesSvc, addressesSvc, err := services(monitor, &desiredKind.Spec, orbID, providerID)
				if err != nil {
					return err
				}

				return destroy(
					machinesSvc,
					addressesSvc,
					&desiredKind.Spec,
				)
			}, migrate, nil
	}
}

func services(monitor mntr.Monitor, desired *Spec, orbID, providerID string) (*machinesService, *addressesSvc, error) {

	jsonKey := []byte(desired.JSONKey.Value)
	credsOption := option.WithCredentialsJSON(jsonKey)
	key := struct {
		ProjectID string `json:"project_id"`
	}{}
	if err := errors.Wrap(json.Unmarshal(jsonKey, &key), "extracting project id from jsonkey failed"); err != nil {
		return nil, nil, err
	}

	monitor = monitor.WithField("projectID", key.ProjectID)

	return newMachinesService(
			monitor,
			desired,
			orbID,
			providerID,
			key.ProjectID,
			credsOption,
		), newAdressesService(
			monitor,
			orbID,
			providerID,
			key.ProjectID,
			desired.Region,
			credsOption,
		), nil
}
