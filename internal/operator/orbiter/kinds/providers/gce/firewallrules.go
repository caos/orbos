package gce

import (
	"fmt"
	"sort"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

type firewallRulesSvc struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	client          *compute.Service
	machinesService *machinesService
}

func (s *firewallRulesSvc) ensure(loadbalancing []*normalizedLoadbalancer) error {
	gceFirewalls, err := s.client.Firewalls.
		List(s.projectID).
		Filter(fmt.Sprintf("description:(orb=%s;provider=%s*)", s.orbID, s.providerID)).
		Fields("items(description,port,requestPath,selfLink)").
		Do()
	if err != nil {
		return err
	}

	var create []*firewall
createLoop:
	for _, lb := range loadbalancing {
		for _, gceFW := range gceFirewalls.Items {
			if gceFW.Description == lb.firewall.gce.Description {
				if gceFW.Allowed[0].Ports[0] != lb.firewall.gce.Allowed[0].Ports[0] ||
					!stringsEqual(gceFW.TargetTags, lb.firewall.gce.TargetTags) ||
					!stringsEqual(gceFW.SourceRanges, lb.firewall.gce.SourceRanges) {
					if err := operate(
						lb.firewall.log("Patching firewall"),
						s.client.Firewalls.Patch(s.projectID, gceFW.Name, lb.firewall.gce).RequestId(uuid.NewV1().String()).Do,
					); err != nil {
						return err
					}
					lb.firewall.log("Firewall patched")()
				}
				lb.firewall.gce.SelfLink = gceFW.SelfLink
				continue createLoop
			}
		}
		lb.firewall.gce.Name = newName()
		create = append(create, lb.firewall)
	}

	var remove []string
removeLoop:

	for _, gceTp := range gceFirewalls.Items {
		for _, lb := range loadbalancing {
			if gceTp.Description == lb.firewall.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, gceTp.Name)
	}

	for _, firewall := range create {
		if err := operate(
			firewall.log("Creating firewall"),
			s.client.Firewalls.
				Insert(s.projectID, firewall.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		firewall.log("Firewall created")()
	}

	for _, healthcheck := range remove {
		if err := operate(
			removeLog(s.monitor, "firewall", healthcheck, false),
			s.client.Firewalls.
				Delete(s.projectID, healthcheck).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(s.monitor, "firewall", healthcheck, true)()
	}

	return nil
}

func stringsEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	sort.Strings(first)
	sort.Strings(second)
	for idx, f := range first {
		if second[idx] != f {
			return false
		}
	}
	return true
}
