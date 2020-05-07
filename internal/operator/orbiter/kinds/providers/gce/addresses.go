package gce

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type addressesSvc struct {
	monitor      mntr.Monitor
	orbID        string
	providerID   string
	projectID    string
	region       string
	clientOption option.ClientOption
	cache        struct {
		client *compute.Service
	}
}

func newAdressesService(
	monitor mntr.Monitor,
	orbID string,
	providerID string,
	projectID string,
	region string,
	clientOption option.ClientOption,
) *addressesSvc {
	return &addressesSvc{
		monitor:      monitor,
		orbID:        orbID,
		providerID:   providerID,
		projectID:    projectID,
		region:       region,
		clientOption: clientOption,
	}
}

func (a *addressesSvc) ensure(addresses map[string]*infra.Address) (bool, error) {
	client, err := a.client()
	if err != nil {
		return false, err
	}
	gceAddresses, err := client.Addresses.
		List(a.projectID, a.region).
		Filter(fmt.Sprintf("addressType=EXTERNAL AND name:%s-%s-*", a.orbID, a.providerID)).
		Fields("items(name,address)").
		Do()
	if err != nil {
		return false, err
	}

	type ensured struct {
		input  *normalizedAddress
		output *compute.Address
	}

	normalizedAddresses := a.normalize(addresses)

	var create []*ensured
createLoop:
	for _, address := range normalizedAddresses {
		for _, gceAddress := range gceAddresses.Items {
			if gceAddress.Name == address.name {
				continue createLoop
			}
		}
		create = append(create, &ensured{
			input: address,
			output: &compute.Address{
				Name:        address.name,
				AddressType: "EXTERNAL",
			},
		})
	}

	var remove []*compute.Address
removeLoop:
	for _, gceAddress := range gceAddresses.Items {
		for _, addressName := range normalizedAddresses {
			if gceAddress.Name == addressName.name {
				continue removeLoop
			}
		}
		remove = append(remove, gceAddress)
	}

	var changed bool
	for _, address := range create {
		if err := operate(
			a.logAddressOpFunc("Creating external address", address.output),
			client.Addresses.Insert(a.projectID, a.region, address.output).RequestId(uuid.NewV1().String()).Do,
		); err != nil {
			return false, err
		}

		newAddr, err := client.Addresses.Get(a.projectID, a.region, address.output.Name).
			Fields("address").
			Do()
		if err != nil {
			return false, err
		}

		changed = true
		*address.input.ip = newAddr.Address
	}
	for _, address := range remove {
		if err := operate(
			a.logAddressOpFunc("Removing external address", address),
			client.Addresses.Delete(a.projectID, a.region, address.Name).RequestId(uuid.NewV1().String()).Do,
		); err != nil {
			return false, err
		}
	}

	return changed, nil
}

type normalizedAddress struct {
	ip   *string
	name string
}

func (a *addressesSvc) client() (client *compute.Service, err error) {
	if a.cache.client == nil {
		a.cache.client, err = compute.NewService(context.TODO(), a.clientOption)
	}
	return a.cache.client, err
}

func (a *addressesSvc) normalize(addresses map[string]*infra.Address) []*normalizedAddress {

	var normalized []*normalizedAddress
outer:
	for addressName, address := range addresses {
		for _, normal := range normalized {
			if *normal.ip == *address.Location {
				normal.name = fmt.Sprintf("%s-%s", normal.name, addressName)
				address.Location = normal.ip
				continue outer
			}
		}
		normalized = append(normalized, &normalizedAddress{
			ip:   address.Location,
			name: fmt.Sprintf("%s-%s-%s", a.orbID, a.providerID, addressName),
		})
	}
	return normalized
}

func (a *addressesSvc) logAddressOpFunc(msg string, address *compute.Address) func() {
	monitor := a.monitor.WithFields(map[string]interface{}{
		"name": address.Name,
	})
	return func() {
		monitor.Info(msg)
	}
}
