package gce

import (
	"fmt"
	"sort"

	uuid "github.com/satori/go.uuid"
)

func queryFirewall(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, error) {
	gceFirewalls, err := context.client.Firewalls.
		List(context.projectID).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(description,name,allowed,targetTags,sourceRanges)").
		Do()
	if err != nil {
		return nil, err
	}

	var operations []func() error

createLoop:
	for _, lb := range loadbalancing {
		for _, fw := range lb.firewalls {
			for _, gceFW := range gceFirewalls.Items {
				if gceFW.Description == fw.gce.Description {
					if gceFW.Allowed[0].Ports[0] != fw.gce.Allowed[0].Ports[0] ||
						!stringsEqual(gceFW.TargetTags, fw.gce.TargetTags) ||
						!stringsEqual(gceFW.SourceRanges, fw.gce.SourceRanges) {
						operations = append(operations, operateFunc(
							fw.log("Patching firewall", true),
							context.client.Firewalls.Patch(context.projectID, gceFW.Name, fw.gce).RequestId(uuid.NewV1().String()).Do,
							toErrFunc(fw.log("Firewall patched", false)),
						))
					}
					continue createLoop
				}
			}
			fw.gce.Name = newName()
			operations = append(operations, operateFunc(
				fw.log("Creating firewall", true),
				context.client.Firewalls.
					Insert(context.projectID, fw.gce).
					RequestId(uuid.NewV1().String()).
					Do,
				toErrFunc(fw.log("Firewall created", false)),
			))
		}
	}

removeLoop:
	for _, gceTp := range gceFirewalls.Items {
		for _, lb := range loadbalancing {
			for _, fw := range lb.firewalls {
				if gceTp.Description == fw.gce.Description {
					continue removeLoop
				}
			}
		}
		operations = append(operations, removeResourceFunc(context.monitor, "firewall", gceTp.Name, context.client.Firewalls.
			Delete(context.projectID, gceTp.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return operations, nil
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
