package static

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caos/orbos/internal/helpers"

	secret2 "github.com/caos/orbos/pkg/secret"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/mntr"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	monitor    mntr.Monitor
	desired    *DesiredV0
	statusFile string
	onCreate   func(machine infra.Machine, pool string) error
	cache      map[string]cachedMachines
}

func NewMachinesService(
	monitor mntr.Monitor,
	desired *DesiredV0,
	id string) *machinesService {
	return &machinesService{
		monitor,
		desired,
		filepath.Join("/var/orbiter", id),
		nil,
		nil,
	}
}

func (c *machinesService) updateKeys() error {

	pools, err := c.ListPools()
	if err != nil {
		panic(err)
	}

	keys := privateKeys(c.desired.Spec)

	for _, pool := range pools {
		machines, err := c.cachedPool(pool)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			if err := machine.UseKey(keys...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *machinesService) ListPools() ([]string, error) {

	pools := make([]string, 0)

	for key := range c.desired.Spec.Pools {
		pools = append(pools, key)
	}

	return pools, nil
}

func (c *machinesService) List(poolName string) (infra.Machines, error) {
	pool, err := c.cachedPool(poolName)
	if err != nil {
		return nil, err
	}

	return pool.Machines(), nil
}

func (c *machinesService) DesiredMachines(poolName string, instances int) int {
	_, err := c.cachedPool(poolName)
	if err != nil {
		return 0
	}

	return instances
}

func (c *machinesService) Create(poolName string, _ int) (infra.Machines, error) {
	pool, err := c.cachedPool(poolName)
	if err != nil {
		return nil, err
	}

	for _, machine := range pool {
		if !machine.X_active {
			machine.X_active = true
			if err := machine.WriteFile(fmt.Sprintf("/home/orbiter/.ssh/authorized_keys"), bytes.NewReader([]byte(c.desired.Spec.Keys.MaintenanceKeyPublic.Value)), 600); err != nil {
				return nil, err
			}

			if err := machine.WriteFile(c.statusFile, strings.NewReader("active"), 600); err != nil {
				return nil, err
			}

			if err := c.onCreate(machine, poolName); err != nil {
				return nil, err
			}

			return []infra.Machine{machine}, nil
		}
	}

	return nil, mntr.ToUserError(errors.New("no machines left"))
}

func (c *machinesService) cachedPool(poolName string) (cachedMachines, error) {

	specifiedMachines, ok := c.desired.Spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("pool %s does not exist", poolName)
	}

	cache, ok := c.cache[poolName]
	if ok {
		return cache, nil
	}

	keys := privateKeys(c.desired.Spec)

	newCache := make([]*machine, 0)

	initializeMachineFunc := func(spec *Machine) func() error {
		return func() error {
			newM := newMachine(c.monitor, c.statusFile, "orbiter", &spec.ID, string(spec.IP),
				spec.RebootRequired,
				func() {
					spec.RebootRequired = true
				}, func() {
					spec.RebootRequired = false
				},
				spec.ReplacementRequired,
				func() {
					spec.ReplacementRequired = true
				}, func() {
					spec.ReplacementRequired = false
				})

			if err := newM.UseKey(keys...); err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			defer buf.Reset()
			if err := newM.ReadFile(c.statusFile, buf); err != nil {
				// if error, treat as active
			}
			newM.X_active = strings.Contains(buf.String(), "active")
			newCache = append(newCache, newM)
			return nil
		}
	}

	var fanout []func() error
	for _, spec := range specifiedMachines {
		fanout = append(fanout, initializeMachineFunc(spec))
	}

	if err := helpers.Fanout(fanout)(); err != nil {
		return nil, err
	}

	if c.cache == nil {
		c.cache = make(map[string]cachedMachines)
	} else if ok { // TODO: Never true
		c.cache[poolName] = nil
		delete(c.cache, poolName)
	}
	c.cache[poolName] = newCache
	return newCache, nil
}

type cachedMachines []*machine

func (c cachedMachines) Machines() infra.Machines {
	machines := make([]infra.Machine, 0)
	for _, machine := range c {
		if machine.X_active {
			machines = append(machines, machine)
		}
	}
	return machines
}

func privateKeys(spec Spec) [][]byte {
	var privateKeys [][]byte
	toBytes := func(key *secret2.Secret) {
		if key != nil {
			privateKeys = append(privateKeys, []byte(key.Value))
		}
	}
	toBytes(spec.Keys.BootstrapKeyPrivate)
	toBytes(spec.Keys.MaintenanceKeyPrivate)
	return privateKeys
}
