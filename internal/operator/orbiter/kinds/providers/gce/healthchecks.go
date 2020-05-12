package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func ensureHealthchecks(context *context, loadbalancing []*normalizedLoadbalancer) error {
	gceHealthchecks, err := context.client.HttpHealthChecks.
		List(context.projectID).
		Filter(fmt.Sprintf("description:(orb=%s;provider=%s*)", context.orbID, context.providerID)).
		Fields("items(description,port,requestPath,selfLink)").
		Do()
	if err != nil {
		return err
	}

	var create []*healthcheck
createLoop:
	for _, lb := range loadbalancing {

		for _, gceTp := range gceHealthchecks.Items {
			if gceTp.Description == lb.healthcheck.gce.Description {
				if gceTp.Port != lb.healthcheck.gce.Port || gceTp.RequestPath != lb.healthcheck.gce.RequestPath {
					if err := operate(
						lb.healthcheck.log("Patching healthcheck"),
						context.client.HttpHealthChecks.Patch(context.projectID, gceTp.Name, lb.healthcheck.gce).
							RequestId(uuid.NewV1().String()).
							Do,
					); err != nil {
						return err
					}
					lb.healthcheck.log("Healthcheck patched")()
				}

				continue createLoop
			}
		}
		lb.healthcheck.gce.Name = newName()
		create = append(create, lb.healthcheck)
	}

	var remove []string
removeLoop:

	for _, gceHC := range gceHealthchecks.Items {
		for _, lb := range loadbalancing {
			if gceHC.Description == lb.healthcheck.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, gceHC.Name)
	}

	for _, healthcheck := range create {
		if err := operate(
			healthcheck.log("Creating healthcheck"),
			context.client.HttpHealthChecks.
				Insert(context.projectID, healthcheck.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		newHC, err := context.client.HttpHealthChecks.Get(context.projectID, healthcheck.gce.Name).
			Fields("selfLink").
			Do()
		if err != nil {
			return err
		}

		healthcheck.gce.SelfLink = newHC.SelfLink
		healthcheck.log("Healthcheck created")()
	}

	for _, healthcheck := range remove {
		if err := operate(
			removeLog(context.monitor, "healthcheck", healthcheck, false),
			context.client.HttpHealthChecks.
				Delete(context.projectID, healthcheck).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(context.monitor, "healthcheck", healthcheck, true)
	}

	return nil
}
