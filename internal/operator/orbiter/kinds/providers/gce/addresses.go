package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"
)

func ensureAddresses(context *context, loadbalancing []*normalizedLoadbalancer) error {

	addresses := normalizedLoadbalancing(loadbalancing).uniqueAddresses()

	gceAddresses, err := context.client.Addresses.
		List(context.projectID, context.region).
		Filter(fmt.Sprintf("addressType=EXTERNAL AND description:(orb=%s;provider=%s*)", context.orbID, context.providerID)).
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
			context.client.Addresses.
				Insert(context.projectID, context.region, address.gce).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}

		newAddr, err := context.client.Addresses.Get(context.projectID, context.region, address.gce.Name).
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
			removeLog(context.monitor, "external address", address.Address, false),
			context.client.Addresses.
				Delete(context.projectID, context.region, address.Name).
				RequestId(uuid.NewV1().String()).
				Do,
		); err != nil {
			return err
		}
		removeLog(context.monitor, "external address", address.Address, true)()
	}

	return nil
}
