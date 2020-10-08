package cs

import (
	"net/http"
	"strconv"

	"github.com/cloudscale-ch/cloudscale-go-sdk"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

func queryFloatingIPs(context *context, loadbalancing map[string][]*dynamic.VIP, writeTo *Current) ([]func() error, []func() error, error) {

	floatingIPs, err := context.client.FloatingIPs.List(context.ctx, func(r *http.Request) {
		params := r.URL.Query()
		params["orb"] = []string{context.orbID}
		params["provider"] = []string{context.providerID}
	})
	if err != nil {
		return nil, nil, err
	}

	var ensure []func() error
createLoop:
	for hostPool, vips := range loadbalancing {
		for vipIdx := range vips {
			vip := vips[vipIdx]
			for transportIdx := range vip.Transport {
				transport := vip.Transport[transportIdx]
				if writeTo.Current.Ingresses == nil {
					writeTo.Current.Ingresses = make(map[string]*infra.Address)
				}
				writeTo.Current.Ingresses[transport.Name] = &infra.Address{}
				alreadyExists := false
				for floatingIPIdx := range floatingIPs {
					floatingIP := floatingIPs[floatingIPIdx]
					if floatingIP.Tags["pool"] == hostPool && floatingIP.Tags["idx"] == strconv.Itoa(vipIdx) {
						writeTo.Current.Ingresses[transport.Name].Location = floatingIP.IP()
						writeTo.Current.Ingresses[transport.Name].FrontendPort = uint16(transport.FrontendPort)
						writeTo.Current.Ingresses[transport.Name].BackendPort = uint16(transport.BackendPort)
						alreadyExists = true
					}
				}
				if alreadyExists {
					continue createLoop
				}
			}
			ensure = append(ensure, func() error {
				newIP, err := context.client.FloatingIPs.Create(context.ctx, &cloudscale.FloatingIPCreateRequest{
					RegionalResourceRequest: cloudscale.RegionalResourceRequest{},
					TaggedResourceRequest: cloudscale.TaggedResourceRequest{Tags: map[string]string{
						"orb":      context.orbID,
						"provider": context.providerID,
						"pool":     hostPool,
						"idx":      strconv.Itoa(vipIdx),
					}},
					IPVersion:      4,
					Server:         "",
					Type:           "regional",
					PrefixLength:   0,
					ReversePointer: "",
				})
				if err != nil {
					return err
				}
				floatingIPs = append(floatingIPs, *newIP)
				return nil
			})
		}
	}

	var remove []func() error

removeLoop:
	for floatingIPIdx := range floatingIPs {
		floatingIP := floatingIPs[floatingIPIdx]
		for hostPool, vips := range loadbalancing {
			for vipIdx := range vips {
				if floatingIP.Tags["pool"] == hostPool && floatingIP.Tags["idx"] == strconv.Itoa(vipIdx) {
					continue removeLoop
				}
			}
		}
		remove = append(remove, func() error {
			return context.client.FloatingIPs.Delete(context.ctx, floatingIP.IP())
		})
	}
	return ensure, remove, nil
}
