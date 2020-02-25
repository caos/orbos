package static

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/mntr"
)

type machinesService struct {
	monitor           mntr.Monitor
	desired           *DesiredV0
	bootstrapkey      []byte
	maintenancekey    []byte
	maintenancekeyPub []byte
	statusFile        string
	desireHostname    func(machine infra.Machine, pool string) error
}

// TODO: Dont accept the whole spec. Accept exactly the values needed (check other constructors too)
func NewMachinesService(
	monitor mntr.Monitor,
	desired *DesiredV0,
	bootstrapkey []byte,
	maintenancekey []byte,
	maintenancekeyPub []byte,
	id string,
	desireHostname func(machine infra.Machine, pool string) error) core.MachinesService {
	return &machinesService{
		monitor,
		desired,
		bootstrapkey,
		maintenancekey,
		maintenancekeyPub,
		filepath.Join("/var/orbiter", id),
		desireHostname,
	}
}

func (c *machinesService) ListPools() ([]string, error) {

	pools := make([]string, 0)

	for key := range c.desired.Spec.Pools {
		pools = append(pools, key)
	}

	return pools, nil
}

func (c *machinesService) List(poolName string, active bool) (infra.Machines, error) {

	cmps, ok := c.desired.Spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	machines := make([]infra.Machine, 0)
	for _, cmp := range cmps {
		var buf bytes.Buffer
		machine := newMachine(c.monitor, c.statusFile, c.desired.Spec.RemoteUser, &cmp.ID, string(cmp.IP))
		if err := machine.UseKey(c.maintenancekey, c.bootstrapkey); err != nil {
			return nil, err
		}
		machine.ReadFile(c.statusFile, &buf)
		isActive := strings.Contains(buf.String(), "active")
		if active && isActive || !active && !isActive {
			machines = append(machines, machine)
		}
		buf.Reset()
	}
	return machines, nil
}

func (c *machinesService) Create(poolName string) (infra.Machine, error) {
	cmps, ok := c.desired.Spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	for _, cmp := range cmps {
		var buf bytes.Buffer
		machine := newMachine(c.monitor, c.statusFile, c.desired.Spec.RemoteUser, &cmp.ID, string(cmp.IP))

		if err := machine.UseKey(c.maintenancekey, c.bootstrapkey); err != nil {
			return nil, err
		}
		machine.ReadFile(c.statusFile, &buf)

		if len(c.maintenancekeyPub) == 0 {
			panic("no maintenancekey")
		}
		if err := machine.WriteFile(c.desired.Spec.RemotePublicKeyPath, bytes.NewReader(c.maintenancekeyPub), 600); err != nil {
			return nil, err
		}

		if strings.Contains(buf.String(), "active") {
			continue
		}

		if err := machine.WriteFile(c.statusFile, strings.NewReader("active"), 600); err != nil {
			return nil, err
		}

		return machine, c.desireHostname(machine, poolName)
	}

	return nil, errors.New("No machines left")
}
