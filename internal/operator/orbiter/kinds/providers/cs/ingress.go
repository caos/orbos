package cs

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

func desiredToCurrentVIP(current *Current) func(vip *dynamiclbmodel.VIP) string {
	return func(vip *dynamic.VIP) string {
		for idx := range vip.Transport {
			transport := vip.Transport[idx]
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	}
}

func notifyMaster(hostPools map[string][]*dynamiclbmodel.VIP, current *Current, context *context) func(m infra.Machine, peers infra.Machines, vips []*dynamiclbmodel.VIP) string {
	return func(m infra.Machine, peers infra.Machines, vips []*dynamiclbmodel.VIP) string {

		seen := make(map[string]bool)
		var hostedVIPs []string
		for hostPool, vips := range hostPools {
			if m.(*machine).server.Tags["pool"] != hostPool {
				continue
			}
			for vipIdx := range vips {
				vip := vips[vipIdx]
				for transpIdx := range vip.Transport {
					transp := vip.Transport[transpIdx]
					addr, ok := current.Current.Ingresses[transp.Name]
					if !ok || addr == nil {
						continue
					}
					if _, ok := seen[addr.Location]; !ok {
						seen[addr.Location] = true
						hostedVIPs = append(hostedVIPs, addr.Location)
					}
				}
			}
		}

		return fmt.Sprintf(`#!/bin/sh

set -e

# API token (with write permission) to access the cloudscale.ch API.
api_token='%s'

# Set of Floating IPs shared between the servers within the same VRRP group.
floating_ipv4='%s'

# UUID of the server that this script is running on.
# The UUID of the server can be retrieved using the API.
server_uuid='%s'

# Call the cloudscale.ch API to assign a specific Floating IP to this server.
set_master() {
	curl \
		-f \
		-H "Authorization: Bearer $api_token" \
		-F server="$server_uuid" \
		"https://api.cloudscale.ch/v1/floating-ips/$1"
}

for VIP in "$floating_ipv4"; do
		set_master $VIP
done
`, context.desired.APIToken.Value, strings.Join(hostedVIPs, " "), m.(*machine).server.UUID)
	}
}
