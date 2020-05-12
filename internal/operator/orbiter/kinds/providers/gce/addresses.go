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

func (a *addressesSvc) query(loadbalancing []*normalizedLoadbalancing) (func() error, error) {

	gceAddresses, err := a.client.Addresses.
		List(a.projectID, a.region).
		Filter(fmt.Sprintf("addressType=EXTERNAL AND name:%s-%s-*", a.orbID, a.providerID)).
		Fields("items(name,address)").
		Do()
	if err != nil {
		return nil, err
	}

	var create []*compute.Address
createLoop:
	for _, address := range loadbalancing {
		for _, gceAddress := range gceAddresses.Items {
			if gceAddress.Address == address.ip {
				continue createLoop
			}
		}

		create = append(create, &compute.Address{
			Name:        fmt.Sprintf("orbos-%s", uuid.NewV1().String()),
			AddressType: "EXTERNAL",
		})
	}

	var remove []*compute.Address
removeLoop:
	for _, gceAddress := range gceAddresses.Items {
		for _, addressName := range loadbalancing {
			if gceAddress.Address == addressName.ip {
				continue removeLoop
			}
		}
		remove = append(remove, gceAddress)
	}

	return func() error {
		for _, address := range create {
			if err := operate(
				a.logAddressOpFunc("Creating external address", address),
				a.client.Addresses.
					Insert(a.projectID, a.region, address).
					RequestId(uuid.NewV1().String()).
					Do,
			); err != nil {
				return err
			}

			newAddr, err := a.client.Addresses.Get(a.projectID, a.region, address.Name).
				Fields("address").
				Do()
			if err != nil {
				return err
			}

			address.Address = newAddr.Address
		}

		for _, address := range remove {
			if err := operate(
				a.logAddressOpFunc("Removing external address", address),
				a.client.Addresses.
					Delete(a.projectID, a.region, address.Name).
					RequestId(uuid.NewV1().String()).
					Do,
			); err != nil {
				return err
			}
		}

		return nil

	}, nil
}

func (a *addressesSvc) logAddressOpFunc(msg string, address *compute.Address) func() {
	monitor := a.monitor.WithFields(map[string]interface{}{
		"name": address.Name,
	})
	return func() {
		monitor.Info(msg)
	}
}
