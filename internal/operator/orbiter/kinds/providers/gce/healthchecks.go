package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func queryHealthchecks(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, error) {
	gceHealthchecks, err := context.client.HttpHealthChecks.
		List(context.projectID).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(description,name,port,requestPath,selfLink)").
		Do()
	if err != nil {
		return nil, err
	}

	var operations []func() error

createLoop:
	for _, lb := range loadbalancing {

		for _, gceHC := range gceHealthchecks.Items {
			if gceHC.Description == lb.healthcheck.gce.Description {
				lb.healthcheck.gce.SelfLink = gceHC.SelfLink
				if gceHC.Port != lb.healthcheck.gce.Port || gceHC.RequestPath != lb.healthcheck.gce.RequestPath {
					operations = append(operations, operateFunc(
						lb.healthcheck.log("Patching healthcheck", true),
						context.client.HttpHealthChecks.Patch(context.projectID, gceHC.Name, lb.healthcheck.gce).
							RequestId(uuid.NewV1().String()).
							Do,
						toErrFunc(lb.healthcheck.log("Healthcheck patched", false)),
					))
				}

				continue createLoop
			}
		}
		lb.healthcheck.gce.Name = newName()
		operations = append(operations, operateFunc(
			lb.healthcheck.log("Creating healthcheck", true),
			context.client.HttpHealthChecks.
				Insert(context.projectID, lb.healthcheck.gce).
				RequestId(uuid.NewV1().String()).
				Do,
			func(hc *healthcheck) func() error {
				return func() error {
					newHC, newHCErr := context.client.HttpHealthChecks.Get(context.projectID, hc.gce.Name).
						Fields("selfLink").
						Do()
					if newHCErr != nil {
						return newHCErr
					}

					hc.gce.SelfLink = newHC.SelfLink
					hc.log("Healthcheck created", false)()
					return nil
				}
			}(lb.healthcheck),
		))
	}

removeLoop:

	for _, gceHC := range gceHealthchecks.Items {
		for _, lb := range loadbalancing {
			if gceHC.Description == lb.healthcheck.gce.Description {
				continue removeLoop
			}
		}
		operations = append(operations, removeResourceFunc(context.monitor, "healthcheck", gceHC.Name, context.client.HttpHealthChecks.
			Delete(context.projectID, gceHC.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return operations, nil
}
