package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

type targetPoolsSvc struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	region          string
	client          *compute.Service
	machinesService *machinesService
}

func (t *targetPoolsSvc) ensure(loadbalancing []*normalizedLoadbalancing) error {
	gcePools, err := t.client.TargetPools.
		List(t.projectID, t.region).
		Filter(fmt.Sprintf("description:%s-%s-*", t.orbID, t.providerID)).
		Fields("items(name,address)").
		Do()
	if err != nil {
		return err
	}

	allInstances, err := t.machinesService.instances()
	if err != nil {
		return err
	}

	type ensurableTargetPool struct {
		instances instances
		gce       *compute.TargetPool
		desired   *targetPool
	}

	var create []*ensurableTargetPool
createLoop:
	for _, ip := range loadbalancing {
		for _, tp := range ip.targetPools {
			poolInstances := allInstances[tp.destPool]

			for _, gceTp := range gcePools.Items {
				if gceTp.Description == tp.id {

					var addInstances []*instance
				addInstanceLoop:
					for _, instance := range poolInstances {
						for _, tpInstance := range gceTp.Instances {
							if instance.id == tpInstance {
								continue addInstanceLoop
							}
						}
						addInstances = append(addInstances, instance)
					}

					if len(addInstances) > 0 {
						richAddInstances := instances(addInstances)
						if err := operate(
							func() {
								t.monitor.
									WithField("instances", richAddInstances.strings(func(i *instance) string { return i.id })).
									Info("Adding instances to target pool")
							},
							t.client.TargetPools.AddInstance(t.projectID, t.region, gceTp.Name, &compute.TargetPoolsAddInstanceRequest{Instances: richAddInstances.refs()}).Do,
						); err != nil {
							return err
						}
					}

					continue createLoop
				}
			}

			create = append(create, &ensurableTargetPool{
				gce: &compute.TargetPool{
					Name:        fmt.Sprintf("orbos-%s", uuid.NewV1().String()),
					Description: tp.id,
					Instances:   instances(poolInstances).strings(func(i *instance) string { return i.url }),
				},
				instances: poolInstances,
				desired:   tp,
			})
		}
	}

	var remove []*ensurableTargetPool
removeLoop:

	for _, gceTp := range gcePools.Items {
		for _, ip := range loadbalancing {
			for _, tp := range ip.targetPools {
				if tp.id == gceTp.Description {
					continue removeLoop
				}
			}
		}
		remove = append(remove, &ensurableTargetPool{
			gce: gceTp,
		})
	}

	for _, targetPool := range create {
		if err := operate(
			t.logOpFunc("Creating target pool", targetPool.desired.id, targetPool.desired, targetPool.instances),
			t.client.TargetPools.
				Insert(t.projectID, t.region, targetPool.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
	}

	for _, targetPool := range remove {
		if err := operate(
			t.logOpFunc("Removing target pool", targetPool.gce.Description, nil, nil),
			t.client.TargetPools.
				Delete(t.projectID, t.region, targetPool.gce.Name).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
	}

	return nil
}

func (t *targetPoolsSvc) logOpFunc(msg string, id string, pool *targetPool, instances instances) func() {
	monitor := t.monitor.WithField("id", id)

	if pool != nil {
		monitor = monitor.WithFields(map[string]interface{}{
			"pool": pool.destPool,
			"name": pool.transport,
		})
	}

	if len(instances) > 0 {
		monitor = monitor.WithField("instances", instances.strings(func(i *instance) string { return i.id }))
	}
	return func() {
		monitor.Info(msg)
	}
}
