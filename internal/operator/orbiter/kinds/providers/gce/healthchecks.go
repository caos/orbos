package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

type healthchecksSvc struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	client          *compute.Service
	machinesService *machinesService
}

func (s *healthchecksSvc) ensure(loadbalancing []*normalizedLoadbalancer) error {
	gceHealthchecks, err := s.client.HttpHealthChecks.
		List(s.projectID).
		Filter(fmt.Sprintf("description:(orb=%s;provider=%s*)", s.orbID, s.providerID)).
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
						s.client.HttpHealthChecks.Patch(s.projectID, gceTp.Name, lb.healthcheck.gce).
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
			s.client.HttpHealthChecks.
				Insert(s.projectID, healthcheck.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		newHC, err := s.client.HttpHealthChecks.Get(s.projectID, healthcheck.gce.Name).
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
			removeLog(s.monitor, "healthcheck", healthcheck, false),
			s.client.HttpHealthChecks.
				Delete(s.projectID, healthcheck).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(s.monitor, "healthcheck", healthcheck, true)
	}

	return nil
}
