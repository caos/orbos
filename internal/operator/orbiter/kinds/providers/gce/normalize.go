package gce

import (
	"fmt"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

type normalizedLoadbalancing struct {
	ip             string
	forwardingRule *forwardingRule
	targetPools    []*targetPool
}

type forwardingRule struct {
}

type externalIP struct {
	ip   string
	name string
}

type targetPool struct {
	id          string
	transport   string
	destPool    string
	healthcheck *healthcheck
}

type healthcheck struct{}

func normalize(spec map[string][]*dynamic.VIP, orbID, providerID string) []*normalizedLoadbalancing {
	var normalized []*normalizedLoadbalancing

	for _, vips := range spec {
		for _, vip := range vips {
			addVIP := true
			for _, src := range vip.Transport {
				if addVIP {
					for _, normal := range normalized {
						if normal.externalIP.ip == *vip.IP {
							normal.externalIP.name = fmt.Sprintf("%s-%s", normal.externalIP.name, src.Name)
							addVIP = false
						}
					}
				}
				if addVIP {
					normalized = append(normalized, &normalizedLoadbalancing{
						externalIP: &externalIP{
							ip:   *vip.IP,
							name: fmt.Sprintf("%s-%s-%s", orbID, providerID, src.Name),
						},
					})
				}
				for _, dest := range src.Destinations {
					// Targetpool ID
					fmt.Sprintf("%s-%s-%s-%s", orbID, providerID, src.Name, dest.Pool)
				}
			}
		}
	}
}
