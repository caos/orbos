package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"
)

func ensureTargetPools(context *context, loadbalancing []*normalizedLoadbalancer) error {
	gcePools, err := context.client.TargetPools.
		List(context.projectID, context.region).
		Filter(fmt.Sprintf("description:(orb=%s;provider=%s*)", context.orbID, context.providerID)).
		Fields("items(description,instances,selfLink,name)").
		Do()
	if err != nil {
		return err
	}

	allInstances, err := context.machinesService.instances()
	if err != nil {
		return err
	}

	type creatableTargetPool struct {
		instances  instances
		targetPool *targetPool
	}

	var create []*creatableTargetPool
createLoop:
	for _, lb := range loadbalancing {
		var poolInstances []*instance
		for _, destPool := range lb.targetPool.destPools {
			poolInstances = append(poolInstances, allInstances[destPool]...)
		}
		for _, gceTp := range gcePools.Items {
			if gceTp.Description == lb.targetPool.gce.Description {

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
						lb.targetPool.log("Adding instances to target pool", richAddInstances),
						context.client.TargetPools.
							AddInstance(
								context.projectID,
								context.region,
								gceTp.Name,
								&compute.TargetPoolsAddInstanceRequest{Instances: richAddInstances.refs()},
							).
							RequestId(uuid.NewV1().String()).
							Do,
					); err != nil {
						return err
					}
					lb.targetPool.log("Instances added to target pool", richAddInstances)()
				}

				if gceTp.HealthChecks[0] != lb.healthcheck.gce.SelfLink {
					if err := operate(
						lb.targetPool.log("Removing healthcheck", nil),
						context.client.TargetPools.RemoveHealthCheck(
							context.projectID,
							context.region,
							gceTp.Name,
							&compute.TargetPoolsRemoveHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: gceTp.HealthChecks[0]}}},
						).
							RequestId(uuid.NewV1().String()).
							Do,
					); err != nil {
						return err
					}
					lb.targetPool.log("Healthcheck removed", nil)()
					if err := operate(
						lb.targetPool.log("Adding healthcheck", nil),
						context.client.TargetPools.AddHealthCheck(
							context.projectID,
							context.region,
							gceTp.Name,
							&compute.TargetPoolsAddHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: lb.healthcheck.gce.SelfLink}}},
						).
							RequestId(uuid.NewV1().String()).
							Do,
					); err != nil {
						return err
					}
					lb.targetPool.log("Healthcheck added", nil)()
				}
				lb.targetPool.gce = gceTp
				continue createLoop
			}
		}

		lb.targetPool.gce.Name = newName()
		lb.targetPool.gce.HealthChecks = []string{lb.healthcheck.gce.SelfLink}
		lb.targetPool.gce.Instances = instances(poolInstances).strings(func(i *instance) string { return i.url })

		create = append(create, &creatableTargetPool{
			instances:  poolInstances,
			targetPool: lb.targetPool,
		})
	}

	var remove []string
removeLoop:

	for _, gceTp := range gcePools.Items {
		for _, lb := range loadbalancing {
			if gceTp.Description == lb.targetPool.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, gceTp.Name)
	}

	for _, targetPool := range create {
		if err := operate(
			targetPool.targetPool.log("Creating target pool", targetPool.instances),
			context.client.TargetPools.
				Insert(context.projectID, context.region, targetPool.targetPool.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		newTP, err := context.client.TargetPools.Get(context.projectID, context.region, targetPool.targetPool.gce.Name).
			Fields("selfLink").
			Do()
		if err != nil {
			return err
		}

		targetPool.targetPool.gce.SelfLink = newTP.SelfLink
		targetPool.targetPool.log("Target pool created", targetPool.instances)()
	}

	for _, targetPool := range remove {
		if err := operate(
			removeLog(context.monitor, "target pool", targetPool, false),
			context.client.TargetPools.
				Delete(context.projectID, context.region, targetPool).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(context.monitor, "target pool", targetPool, true)()
	}

	return nil
}
