package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

var _ ensureLBFunc = queryAddresses

func queryAddresses(context *context, loadbalancing []*normalizedLoadbalancer) ([]func() error, []func() error, error) {

	addresses := normalizedLoadbalancing(loadbalancing).uniqueAddresses()

	gceAddresses, err := context.client.Addresses.
		List(context.projectID, context.desired.Region).
		Filter(fmt.Sprintf(`description : "orb=%s;provider=%s*"`, context.orbID, context.providerID)).
		Fields("items(address,name,description)").
		Do()
	if err != nil {
		return nil, nil, err
	}

	var ensure []func() error

createLoop:
	for _, addr := range addresses {
		for _, gceAddress := range gceAddresses.Items {
			if gceAddress.Description == addr.gce.Description {
				addr.gce.Address = gceAddress.Address
				continue createLoop
			}
		}

		addr.gce.Name = newName()
		ensure = append(ensure, operateFunc(
			addr.log("Creating external address", true),
			computeOpCall(context.client.Addresses.
				Insert(context.projectID, context.desired.Region, addr.gce).
				RequestId(uuid.NewV1().String()).
				Do),
			func(a *address) func() error {
				return func() error {
					newAddr, newAddrErr := context.client.Addresses.Get(context.projectID, context.desired.Region, a.gce.Name).
						Fields("address").
						Do()
					if newAddrErr != nil {
						return newAddrErr
					}
					a.gce.Address = newAddr.Address
					a.log("External address created", false)()
					return nil
				}
			}(addr)))
	}

	var remove []func() error
removeLoop:
	for _, gceAddress := range gceAddresses.Items {
		for _, address := range addresses {
			if gceAddress.Description == address.gce.Description {
				continue removeLoop
			}
		}
		remove = append(remove, removeResourceFunc(context.monitor, "external address", gceAddress.Name, context.client.Addresses.
			Delete(context.projectID, context.desired.Region, gceAddress.Name).
			RequestId(uuid.NewV1().String()).
			Do))
	}
	return ensure, remove, nil
}
