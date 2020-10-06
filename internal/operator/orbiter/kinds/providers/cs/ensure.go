package cs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudscale-ch/cloudscale-go-sdk"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbos/internal/api"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
	nodeAgentsCurrent *common.CurrentNodeAgents,
	nodeAgentsDesired *common.DesiredNodeAgents,
	naFuncs core.IterateNodeAgentFuncs,
	orbiterCommit string,
) (ensureFunc orbiter.EnsureFunc, err error) {

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		panic(errors.Errorf("Unknown or unsupported load balancing of type %T", lb))
	}

	floatingIPs, err := context.client.FloatingIPs.List(context.ctx, func(r *http.Request) {
		params := r.URL.Query()
		params["orb"] = []string{context.orbID}
		params["provider"] = []string{context.providerID}
	})
	if err != nil {
		return nil, err
	}

	hostPools, err := lbCurrent.Current.Spec(context.machinesService)
	if err != nil {
		return nil, err
	}

	var ensureFloatingIPs []func() error
	current.Current.Ingresses = make(map[string]*infra.Address)
	for _, floatingIP := range floatingIPs {
		for hostPool, vips := range hostPools {
			if floatingIP.Tags["pool"] == hostPool {
				vipFound := false
				for idx, vip := range vips {
					if floatingIP.Tags["idx"] == strconv.FormatInt(int64(idx), 10) {
						vipFound = true
						for _, transport := range vip.Transport {
							current.Current.Ingresses[transport.Name] = &infra.Address{
								Location:     floatingIP.IP(),
								FrontendPort: uint16(transport.FrontendPort),
								BackendPort:  uint16(transport.BackendPort),
							}
						}
						break
					}
					ensureFloatingIPs = append(ensureFloatingIPs, func() error {
						return context.client.FloatingIPs.Delete(context.ctx, floatingIP.IP())
					})
				}
			}
		}
	}

	var externalPools []string
	for hostPool, vips := range hostPools {
		externalPools = append(externalPools, hostPool)
		for _, floatingIP := range floatingIPs {
			if floatingIP.Tags["pool"] == hostPool {
				for idx, vip := range vips {
					if floatingIP.Tags["idx"] == strconv.FormatInt(int64(idx), 10) {
						for _, transport := range vip.Transport {
							current.Current.Ingresses[transport.Name] = &infra.Address{
								Location:     floatingIP.IP(),
								FrontendPort: uint16(transport.FrontendPort),
								BackendPort:  uint16(transport.BackendPort),
							}
						}
						break
					}
				}
			}
		}
	}

	queryNA, installNA := naFuncs(nodeAgentsCurrent)

	ensureNodeAgent := func(pool string, machine infra.Machine) error {
		running, err := queryNA(machine, orbiterCommit)
		if err != nil {
			return err
		}
		if !running {
			return installNA(machine)
		}

		return nil
	}

	context.machinesService.onCreate = ensureNodeAgent
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, true, func(machine infra.Machine, peers infra.Machines, vips []*dynamiclbmodel.VIP) string {
		return ""
	}, func(vip *dynamic.VIP) string {
		for _, transport := range vip.Transport {
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	})
	return func(pdf api.PushDesiredFunc) *orbiter.EnsureResult {
		var done bool
		return orbiter.ToEnsureResult(done, helpers.Fanout([]func() error{
			func() error { return helpers.Fanout(ensureFloatingIPs)() },
			func() error {
				pools, err := context.machinesService.instances()
				if err != nil {
					return err
				}

				var ensureNodeAgents []func() error
				var ensurePublicInterfaces []func() error
				for pool, machines := range pools {
					isExternal := context.machinesService.oneoff
					for _, externalPool := range externalPools {
						if externalPool == pool {
							isExternal = true
						}
					}
					for _, machine := range machines {
						hasPublicInterface := false
						var privateInterfaces []cloudscale.Interface
						for _, interf := range machine.server.Interfaces {
							if interf.Type == "public" {
								hasPublicInterface = true
							} else {
								privateInterfaces = append(privateInterfaces, interf)
							}
						}

						var updateInterfaces []cloudscale.InterfaceRequest
						if isExternal && !hasPublicInterface {
							updateInterfaces = append(interfaces(machine.server.Interfaces).toRequests(), cloudscale.InterfaceRequest{Network: "public"})
						}

						if !isExternal && hasPublicInterface {
							updateInterfaces = interfaces(privateInterfaces).toRequests()
						}

						if updateInterfaces != nil {
							ensurePublicInterfaces = append(ensurePublicInterfaces, updateServerFunc(context, machine.server, &updateInterfaces))
						}

						ensureNodeAgents = append(ensureNodeAgents, func(p string, m infra.Machine) func() error {
							return func() error {
								return ensureNodeAgent(p, m)
							}
						}(pool, machine))
					}
				}
				return helpers.Fanout(ensureNodeAgents)()
			},
			func() error {
				var err error
				done, err = wrappedMachines.InitializeDesiredNodeAgents()
				return err
			},
		})())
	}, addPools(current, desired, wrappedMachines)
}

func updateServerFunc(context *context, server *cloudscale.Server, interfaces *[]cloudscale.InterfaceRequest) func() error {
	return func() error {
		return context.client.Servers.Update(context.ctx, server.UUID, &cloudscale.ServerUpdateRequest{
			TaggedResourceRequest: cloudscale.TaggedResourceRequest{Tags: server.Tags},
			Name:                  server.Name,
			Status:                server.Status,
			Flavor:                server.Flavor.Name,
			Interfaces:            interfaces,
		})
	}

}

type interfaces []cloudscale.Interface

func (i interfaces) toRequests() []cloudscale.InterfaceRequest {
	var requests []cloudscale.InterfaceRequest
	for _, interf := range i {
		addr := addresses(interf.Addresses).toRequest()
		requests = append(requests, cloudscale.InterfaceRequest{
			Network:   interf.Network.UUID,
			Addresses: &addr,
		})
	}
	return requests
}

type addresses []cloudscale.Address

func (a addresses) toRequest() []cloudscale.AddressRequest {
	var requests []cloudscale.AddressRequest
	for _, addr := range a {
		requests = append(requests, cloudscale.AddressRequest{
			Subnet:  addr.Subnet.UUID,
			Address: addr.Address,
		})
	}
	return requests
}
