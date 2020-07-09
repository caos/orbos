package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

type MachinesFunc func(monitor mntr.Monitor, desiredTree *tree.Tree, repoURL string) (machineIDs []string, machines map[string]infra.Machine, err error)

func ListMachines() MachinesFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, repoURL string) (machineIDs []string, machines map[string]infra.Machine, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		machines = make(map[string]infra.Machine, 0)
		machineIDs = make([]string, 0)

		for provID, providerTree := range desiredKind.Providers {
			providerMachines, err := providers.ListMachines(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				providerTree,
				provID,
				repoURL,
			)
			if err != nil {
				return nil, nil, err
			}

			for id, providerMachine := range providerMachines {
				machineID := provID + "." + id
				machineIDs = append(machineIDs, machineID)
				machines[machineID] = providerMachine
			}
		}
		return machineIDs, machines, nil
	}
}
