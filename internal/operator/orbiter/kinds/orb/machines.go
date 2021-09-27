package orb

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/providers"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/labels"
	"github.com/caos/orbos/v5/pkg/tree"
)

type MachinesFunc func(monitor mntr.Monitor, desiredTree *tree.Tree, orbID string) (machineIDs []string, machines map[string]infra.Machine, err error)

func ListMachines(operarorLabels *labels.Operator) MachinesFunc {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, orbID string) (machineIDs []string, machines map[string]infra.Machine, err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("building %s failed: %w", desiredTree.Common.Kind, err)
			}
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing desired state failed: %w", err)
		}
		desiredTree.Parsed = desiredKind

		machines = make(map[string]infra.Machine, 0)
		machineIDs = make([]string, 0)

		for clusterID, clusterTree := range desiredKind.Clusters {
			clusterCurrent := &tree.Tree{}
			_, _, _, _, _, err := clusters.GetQueryAndDestroyFuncs(
				monitor,
				operarorLabels,
				clusterID,
				clusterTree,
				true,
				false,
				false,
				clusterCurrent,
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				return nil, nil, err
			}
		}

		for provID, providerTree := range desiredKind.Providers {

			providerMachines, err := providers.ListMachines(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				providerTree,
				provID,
				orbID,
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
