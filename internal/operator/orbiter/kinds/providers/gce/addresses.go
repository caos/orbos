package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

type addressesSvc struct {
	monitor    mntr.Monitor
	orbID      string
	providerID string
	projectID  string
	region     string
	client     *compute.Service
}

func newAdressesService(
	monitor mntr.Monitor,
	orbID string,
	providerID string,
	projectID string,
	region string,
	client *compute.Service,
) *addressesSvc {
	return &addressesSvc{
		monitor:    monitor,
		orbID:      orbID,
		providerID: providerID,
		projectID:  projectID,
		region:     region,
		client:     client,
	}
}

func (s *addressesSvc) ensure(loadbalancing []*normalizedLoadbalancer) error {

	addresses := normalizedLoadbalancing(loadbalancing).uniqueAddresses()

	gceAddresses, err := s.client.Addresses.
		List(s.projectID, s.region).
		Filter(fmt.Sprintf("addressType=EXTERNAL AND description:(orb=%s;provider=%s*)", s.orbID, s.providerID)).
		Fields("items(address,description)").
		Do()
	if err != nil {
		return err
	}

	var create []*address
createLoop:
	for _, address := range addresses {
		for _, gceAddress := range gceAddresses.Items {
			if gceAddress.Description == address.gce.Description {
				address.gce = gceAddress
				continue createLoop
			}
		}

		address.gce.Name = newName()
		create = append(create, address)
	}

	var remove []*compute.Address
removeLoop:
	for _, gceAddress := range gceAddresses.Items {
		for _, address := range addresses {
			if gceAddress.Description == address.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, gceAddress)
	}

	for _, address := range create {
		if err := operate(
			address.log("Creating external address"),
			s.client.Addresses.
				Insert(s.projectID, s.region, address.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		newAddr, err := s.client.Addresses.Get(s.projectID, s.region, address.gce.Name).
			Fields("address").
			Do()
		if err != nil {
			return err
		}
		address.gce.Address = newAddr.Address
		address.log("External address created")()
	}

	for _, address := range remove {
		if err := operate(
			removeLog(s.monitor, "external address", address.Address, false),
			s.client.Addresses.
				Delete(s.projectID, s.region, address.Name).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(s.monitor, "external address", address.Address, true)()
	}

	return nil
}
