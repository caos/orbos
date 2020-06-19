package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

var _ ensureLBFunc = queryForwardingRules

func queryForwardingRules(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, []func() error, error) {
	gceRules, err := context.client.ForwardingRules.
		List(context.projectID, context.desired.Region).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(description,name,target,portRange,IPAddress)").
		Do()
	if err != nil {
		return nil, nil, err
	}

	var ensure []func() error
	assignRefs := func(lb *normalizedLoadbalancer) {
		lb.forwardingRule.gce.Target = lb.targetPool.gce.SelfLink
		lb.forwardingRule.gce.IPAddress = lb.address.gce.Address
	}

createLoop:
	for _, lb := range loadbalancing {
		for _, gceRule := range gceRules.Items {
			if gceRule.Description == lb.forwardingRule.gce.Description {
				assignRefs(lb)
				lb.forwardingRule.gce.Name = gceRule.Name
				if gceRule.Target != lb.forwardingRule.gce.Target || gceRule.PortRange != lb.forwardingRule.gce.PortRange || gceRule.IPAddress != lb.forwardingRule.gce.IPAddress {
					ensure = append(ensure, operateFunc(
						lb.forwardingRule.log("Patching forwarding rule", true),
						computeOpCall(context.client.ForwardingRules.Patch(context.projectID, context.desired.Region, gceRule.Name, lb.forwardingRule.gce).
							RequestId(uuid.NewV1().String()).
							Do),
						toErrFunc(lb.forwardingRule.log("Forwarding rule patched", false)),
					))
				}
				continue createLoop
			}
		}

		lb.forwardingRule.gce.Name = newName()
		ensure = append(ensure, operateFunc(
			func(l *normalizedLoadbalancer) func() {
				return func() {
					assignRefs(l)
					l.forwardingRule.log("Creating forwarding rule", true)()
				}
			}(lb),
			computeOpCall(context.client.ForwardingRules.
				Insert(context.projectID, context.desired.Region, lb.forwardingRule.gce).
				RequestId(uuid.NewV1().String()).
				Do),
			toErrFunc(lb.forwardingRule.log("Forwarding rule created", false)),
		))
	}

	var remove []func() error

removeLoop:
	for _, rule := range gceRules.Items {
		for _, lb := range loadbalancing {
			if rule.Description == lb.forwardingRule.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, removeResourceFunc(
			context.monitor, "forwarding rule", rule.Name, context.client.ForwardingRules.
				Delete(context.projectID, context.desired.Region, rule.Name).
				RequestId(uuid.NewV1().String()).
				Do,
		))
	}
	return ensure, remove, nil
}
