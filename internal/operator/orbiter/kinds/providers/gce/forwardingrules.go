package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

type forwardingRulesSvc struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	region          string
	client          *compute.Service
	machinesService *machinesService
}

func (s *forwardingRulesSvc) ensure(loadbalancing []*normalizedLoadbalancer) error {
	gceRules, err := s.client.ForwardingRules.
		List(s.projectID, s.region).
		Filter(fmt.Sprintf("loadBalancingScheme=EXTERNAL AND description:(orb=%s;provider=%s*)", s.orbID, s.providerID)).
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
						s.client.ForwardingRules.Patch(s.projectID, s.region, gceRule.Name, lb.forwardingRule.gce).
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
			s.client.ForwardingRules.
				Insert(s.projectID, s.region, rule.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		rule.log("Forwarding rule created")()
	}

	for _, forwardingRule := range remove {
		if err := operate(
			removeLog(s.monitor, "forwarding rule", forwardingRule, false),
			s.client.ForwardingRules.
				Delete(s.projectID, s.region, forwardingRule).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(s.monitor, "forwarding rule", forwardingRule, true)()
	}

	return nil
}
