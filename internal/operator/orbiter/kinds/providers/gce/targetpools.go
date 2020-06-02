package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"
)

var _ queryFunc = queryTargetPools

func queryTargetPools(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, []func() error, error) {
	gcePools, err := context.client.TargetPools.
		List(context.projectID, context.desired.Region).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(description,name,instances,selfLink,name)").
		Do()
	if err != nil {
		return nil, nil, err
	}

	assignRefs := func(lb *normalizedLoadbalancer) {
		lb.targetPool.gce.HealthChecks = []string{lb.healthcheck.gce.SelfLink}
	}

	var ensure []func() error

createLoop:
	for _, lb := range loadbalancing {
		for _, gceTp := range gcePools.Items {
			if gceTp.Description == lb.targetPool.gce.Description {
				lb.targetPool.gce.SelfLink = gceTp.SelfLink
				lb.targetPool.gce.Name = gceTp.Name
				assignRefs(lb)
				if len(gceTp.HealthChecks) > 0 && gceTp.HealthChecks[0] != lb.targetPool.gce.HealthChecks[0] {
					ensure = append(ensure, operateFunc(
						lb.targetPool.log("Removing healthcheck", true, nil),
						computeOpCall(context.client.TargetPools.RemoveHealthCheck(
							context.projectID,
							context.desired.Region,
							gceTp.Name,
							&compute.TargetPoolsRemoveHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: gceTp.HealthChecks[0]}}},
						).
							RequestId(uuid.NewV1().String()).
							Do),
						toErrFunc(lb.targetPool.log("Healthcheck removed", false, nil)),
					))

					ensure = append(ensure, operateFunc(
						lb.targetPool.log("Adding healthcheck", true, nil),
						computeOpCall(context.client.TargetPools.AddHealthCheck(
							context.projectID,
							context.desired.Region,
							gceTp.Name,
							&compute.TargetPoolsAddHealthCheckRequest{HealthChecks: []*compute.HealthCheckReference{{HealthCheck: lb.healthcheck.gce.SelfLink}}},
						).
							RequestId(uuid.NewV1().String()).
							Do),
						toErrFunc(lb.targetPool.log("Healthcheck added", false, nil)),
					))
				}
				continue createLoop
			}
		}

		lb.targetPool.gce.Name = newName()

		ensure = append(ensure, operateFunc(
			func(l *normalizedLoadbalancer) func() {
				return func() {
					assignRefs(l)
					lb.targetPool.log("Creating target pool", true, nil)()
				}
			}(lb),
			computeOpCall(context.client.TargetPools.
				Insert(context.projectID, context.desired.Region, lb.targetPool.gce).
				RequestId(uuid.NewV1().String()).
				Do),
			func(pool *targetPool) func() error {
				return func() error {
					newTP, err := context.client.TargetPools.Get(context.projectID, context.desired.Region, pool.gce.Name).
						Fields("selfLink").
						Do()
					if err != nil {
						return err
					}

					pool.gce.SelfLink = newTP.SelfLink
					pool.log("Target pool created", false, nil)()
					return nil
				}
			}(lb.targetPool),
		))
	}

	var remove []func() error
removeLoop:
	for _, gceTp := range gcePools.Items {
		for _, lb := range loadbalancing {
			if gceTp.Description == lb.targetPool.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, removeResourceFunc(context.monitor, "target pool", gceTp.Name, context.client.TargetPools.
			Delete(context.projectID, context.desired.Region, gceTp.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return ensure, remove, nil
}
