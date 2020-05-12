package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func ensureForwardingRules(context *context, loadbalancing []*normalizedLoadbalancer) error {
	gceRules, err := context.client.ForwardingRules.
		List(context.projectID, context.region).
		Filter(fmt.Sprintf("loadBalancingScheme=EXTERNAL AND description:(orb=%s;provider=%s*)", context.orbID, context.providerID)).
		Fields("items(description,target,portRange,selfLink)").
		Do()
	if err != nil {
		return err
	}

	var create []*forwardingRule
createLoop:
	for _, lb := range loadbalancing {
		for _, gceRule := range gceRules.Items {
			if gceRule.Description == lb.forwardingRule.gce.Description {
				if gceRule.Target != lb.targetPool.gce.SelfLink || gceRule.PortRange != lb.forwardingRule.gce.PortRange {
					if err := operate(
						lb.forwardingRule.log("Patching forwarding rule"),
						context.client.ForwardingRules.Patch(context.projectID, context.region, gceRule.Name, lb.forwardingRule.gce).
							RequestId(uuid.NewV1().String()).
							Do,
					); err != nil {
						return err
					}
					lb.forwardingRule.log("Forwarding rule patched")()
				}
				continue createLoop
			}
		}

		lb.forwardingRule.gce.Name = newName()
		lb.forwardingRule.gce.Target = lb.targetPool.gce.SelfLink
		create = append(create, lb.forwardingRule)
	}

	var remove []string
removeLoop:

	for _, gceTp := range gceRules.Items {
		for _, lb := range loadbalancing {
			if gceTp.Description == lb.forwardingRule.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, gceTp.Name)
	}

	for _, rule := range create {
		if err := operate(
			rule.log("Creating forwarding rule"),
			context.client.ForwardingRules.
				Insert(context.projectID, context.region, rule.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		rule.log("Forwarding rule created")()
	}

	for _, forwardingRule := range remove {
		if err := operate(
			removeLog(context.monitor, "forwarding rule", forwardingRule, false),
			context.client.ForwardingRules.
				Delete(context.projectID, context.region, forwardingRule).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(context.monitor, "forwarding rule", forwardingRule, true)()
	}

	return nil
}
