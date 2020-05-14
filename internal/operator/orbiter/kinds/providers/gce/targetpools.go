package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"
)

func queryTargetPools(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, error) {
	gcePools, err := context.client.TargetPools.
		List(context.projectID, context.region).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(description,name,instances,selfLink,name)").
		Do()
	if err != nil {
		return nil, err
	}

	allInstances, err := context.machinesService.instances()
	if err != nil {
		return nil, err
	}

	assignRefs := func(lb *normalizedLoadbalancer) {
		lb.targetPool.gce.HealthChecks = []string{lb.healthcheck.gce.SelfLink}
	}

	var operations []func() error

createLoop:
	for _, lb := range loadbalancing {
		var poolInstances []*instance
		for _, destPool := range lb.targetPool.destPools {
			poolInstances = append(poolInstances, allInstances[destPool]...)
		}
		for _, gceTp := range gcePools.Items {
			if gceTp.Description == lb.targetPool.gce.Description {
				lb.targetPool.gce.SelfLink = gceTp.SelfLink
				assignRefs(lb)
				var addInstances []*instance
			addInstanceLoop:
				for _, instance := range poolInstances {
					for _, tpInstance := range gceTp.Instances {
						if instance.url == tpInstance {
							continue addInstanceLoop
						}
					}
					addInstances = append(addInstances, instance)
				}

				if len(addInstances) > 0 {
					richAddInstances := instances(addInstances)
					operations = append(operations, operateFunc(
						lb.targetPool.log("Adding instances to target pool", true, richAddInstances),
						context.client.TargetPools.
							AddInstance(
								context.projectID,
								context.region,
								gceTp.Name,
								&compute.TargetPoolsAddInstanceRequest{Instances: richAddInstances.refs()},
							).
							RequestId(uuid.NewV1().String()).
							Do,
						toErrFunc(lb.targetPool.log("Instances added to target pool", false, richAddInstances)),
					))
				}

				if len(gceTp.HealthChecks) > 0 && gceTp.HealthChecks[0] != lb.targetPool.gce.HealthChecks[0] {
					operations = append(operations, operateFunc(
						lb.targetPool.log("Removing healthcheck", true, nil),
						context.client.TargetPools.RemoveHealthCheck(
							context.projectID,
							context.region,
							gceTp.Name,
							&compute.TargetPoolsRemoveHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: gceTp.HealthChecks[0]}}},
						).
							RequestId(uuid.NewV1().String()).
							Do,
						toErrFunc(lb.targetPool.log("Healthcheck removed", false, nil)),
					))

					operations = append(operations, operateFunc(
						lb.targetPool.log("Adding healthcheck", true, nil),
						context.client.TargetPools.AddHealthCheck(
							context.projectID,
							context.region,
							gceTp.Name,
							&compute.TargetPoolsAddHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: lb.healthcheck.gce.SelfLink}}},
						).
							RequestId(uuid.NewV1().String()).
							Do,
						toErrFunc(lb.targetPool.log("Healthcheck added", false, nil)),
					))
				}
				continue createLoop
			}
		}

		richInstances := instances(poolInstances)
		lb.targetPool.gce.Name = newName()
		lb.targetPool.gce.Instances = richInstances.strings(func(i *instance) string { return i.url })

		operations = append(operations, operateFunc(
			func(l *normalizedLoadbalancer) func() {
				return func() {
					assignRefs(l)
					lb.targetPool.log("Creating target pool", true, richInstances)()
				}
			}(lb),
			context.client.TargetPools.
				Insert(context.projectID, context.region, lb.targetPool.gce).
				RequestId(uuid.NewV1().String()).
				Do,
			func(pool *targetPool) func() error {
				return func() error {
					newTP, err := context.client.TargetPools.Get(context.projectID, context.region, pool.gce.Name).
						Fields("selfLink").
						Do()
					if err != nil {
						return err
					}

					pool.gce.SelfLink = newTP.SelfLink
					pool.log("Target pool created", false, richInstances)()
					return nil
				}
			}(lb.targetPool),
		))
	}

removeLoop:

	for _, gceTp := range gcePools.Items {
		for _, lb := range loadbalancing {
			if gceTp.Description == lb.targetPool.gce.Description {
				continue removeLoop
			}
		}
		operations = append(operations, removeResourceFunc(context.monitor, "target pool", gceTp.Name, context.client.TargetPools.
			Delete(context.projectID, context.region, gceTp.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return operations, nil
}
