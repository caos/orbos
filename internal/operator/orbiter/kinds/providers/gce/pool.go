package gce

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/compute/v1"
)

var _ infra.Pool = (*infraPool)(nil)

type infraPool struct {
	pool       string
	normalized []*normalizedLoadbalancer
	context    *context
}

func newInfraPool(pool string, context *context, normalized []*normalizedLoadbalancer) *infraPool {
	return &infraPool{
		pool:       pool,
		normalized: normalized,
		context:    context,
	}
}

func (i *infraPool) EnsureMember(machine infra.Machine) error {
	return i.ensureMembers(machine)
}

func (i *infraPool) EnsureMembers() error {
	return i.ensureMembers(nil)
}

func (i *infraPool) ensureMembers(machine infra.Machine) error {

	allInstances, err := i.context.machinesService.instances()
	if err != nil {
		return err
	}

	poolInstances := allInstances[i.pool]

	for _, n := range i.normalized {
	destpoolLoop:
		for _, destPool := range n.targetPool.destPools {
			if destPool == i.pool {

				var addInstances []*instance
			addInstanceLoop:
				for _, instance := range poolInstances {
					if machine != nil && machine.ID() != instance.ID() {
						continue addInstanceLoop
					}
					for _, tpInstance := range n.targetPool.gce.Instances {
						if instance.url == tpInstance {
							continue addInstanceLoop
						}
					}
					addInstances = append(addInstances, instance)
				}

				if len(addInstances) == 0 {
					continue destpoolLoop
				}

				if err := operateFunc(
					n.targetPool.log("Adding instances to target pool", true, addInstances),
					i.context.client.TargetPools.
						AddInstance(
							i.context.projectID,
							i.context.region,
							n.targetPool.gce.Name,
							&compute.TargetPoolsAddInstanceRequest{Instances: instances(addInstances).refs()},
						).
						RequestId(uuid.NewV1().String()).
						Do,
					toErrFunc(n.targetPool.log("Instances added to target pool", false, addInstances)),
				)(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (i *infraPool) GetMachines() (infra.Machines, error) {
	return i.context.machinesService.List(i.pool)
}

func (i *infraPool) AddMachine() (infra.Machine, error) {
	return i.context.machinesService.Create(i.pool)
}
