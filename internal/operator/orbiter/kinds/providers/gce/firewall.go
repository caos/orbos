package gce

import (
	"fmt"
	"sort"

	uuid "github.com/satori/go.uuid"
)

var _ ensureFWFunc = queryFirewall

func queryFirewall(context *svcConfig, firewalls []*firewall) ([]func() error, []func() error, error) {
	gceFirewalls, err := context.client.Firewalls.
		List(context.projectID).
		Filter(fmt.Sprintf(`network = "https://www.googleapis.com/compute/v1/%s"`, context.networkURL)).
		Fields("items(network,name,description,allowed,targetTags,sourceRanges)").
		Do()
	if err != nil {
		return nil, nil, err
	}

	var ensure []func() error
createLoop:
	for _, fw := range firewalls {
		for _, gceFW := range gceFirewalls.Items {
			if fw.gce.Description == gceFW.Description {
				if gceFW.Allowed[0].Ports[0] != fw.gce.Allowed[0].Ports[0] ||
					!stringsEqual(gceFW.TargetTags, fw.gce.TargetTags) ||
					!stringsEqual(gceFW.SourceRanges, fw.gce.SourceRanges) {
					ensure = append(ensure, operateFunc(
						fw.log("Patching firewall", true),
						computeOpCall(context.client.Firewalls.Patch(context.projectID, gceFW.Name, fw.gce).RequestId(uuid.NewV1().String()).Do),
						toErrFunc(fw.log("Firewall patched", false)),
					))
				}
				continue createLoop
			}
		}
		fw.gce.Name = newName()
		ensure = append(ensure, operateFunc(
			fw.log("Creating firewall", true),
			computeOpCall(context.client.Firewalls.
				Insert(context.projectID, fw.gce).
				RequestId(uuid.NewV1().String()).
				Do),
			toErrFunc(fw.log("Firewall created", false)),
		))
	}

	var remove []func() error
removeLoop:
	for _, gceTp := range gceFirewalls.Items {
		for _, fw := range firewalls {
			if gceTp.Description == fw.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, removeResourceFunc(context.monitor, "firewall", gceTp.Name, context.client.Firewalls.
			Delete(context.projectID, gceTp.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return ensure, remove, nil
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
