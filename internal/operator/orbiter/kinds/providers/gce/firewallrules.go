package gce

import (
	"fmt"
	"sort"

	uuid "github.com/satori/go.uuid"
)

func ensureFirewall(context *context, loadbalancing []*normalizedLoadbalancer) error {
	gceFirewalls, err := context.client.Firewalls.
		List(context.projectID).
		Filter(fmt.Sprintf("description:(orb=%s;provider=%s*)", context.orbID, context.providerID)).
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
						context.client.Firewalls.Patch(context.projectID, gceFW.Name, lb.firewall.gce).RequestId(uuid.NewV1().String()).Do,
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
			context.client.Firewalls.
				Insert(context.projectID, firewall.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		firewall.log("Firewall created")()
	}

	for _, healthcheck := range remove {
		if err := operate(
			removeLog(context.monitor, "firewall", healthcheck, false),
			context.client.Firewalls.
				Delete(context.projectID, healthcheck).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(context.monitor, "firewall", healthcheck, true)()
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
