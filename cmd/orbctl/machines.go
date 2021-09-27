package main

import (
	"fmt"

	orbcfg "github.com/caos/orbos/v5/pkg/orb"

	"github.com/caos/orbos/v5/pkg/labels"

	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
	"github.com/caos/orbos/v5/pkg/tree"
)

func machines(monitor mntr.Monitor, gitClient *git.Client, orbConfig *orbcfg.Orb, do func(machineIDs []string, machines map[string]infra.Machine, desired *tree.Tree) error) error {

	if err := initRepo(orbConfig, gitClient); err != nil {
		return err
	}

	if !gitClient.Exists(git.OrbiterFile) {
		return mntr.ToUserError(fmt.Errorf("%s not found", git.OrbiterFile))
	}

	monitor.Debug("Reading machines from orbiter.yml")

	desired, err := gitClient.ReadTree(git.OrbiterFile)
	if err != nil {
		return err
	}

	listMachines := orb.ListMachines(labels.NoopOperator("ORBOS"))

	orbID, err := orbConfig.ID()
	if err != nil {
		return err
	}

	machineIDs, machines, err := listMachines(
		monitor,
		desired,
		orbID,
	)

	if err != nil {
		return err
	}

	return do(machineIDs, machines, desired)
}
