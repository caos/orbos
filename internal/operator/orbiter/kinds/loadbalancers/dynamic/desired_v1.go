package dynamic

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/pkg/tree"
)

type DesiredV1 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   map[string][]*VIPV1
}

func v1tov2(v1 *DesiredV1) *Desired {
	d := Desired{
		Spec: make(map[string][]*VIP, len(v1.Spec)),
	}
	for poolNameV0, poolV0 := range v1.Spec {
		vips := make([]*VIP, len(poolV0))
		for vipIdx, v := range poolV0 {

			var transport []*Transport
			for _, t := range v.Transport {
				var destV1 DestinationV1
				var backendPools []string
				for destIdx, dest := range t.Destinations {
					backendPools = append(backendPools, dest.Pool)
					if destIdx == 0 {
						destV1 = *dest
					}
				}
				if len(t.Destinations) > 0 {
					destV1 = *t.Destinations[0]
				}

				var wl []*orbiter.CIDR
				for _, c := range t.Whitelist {
					cidr := orbiter.CIDR(*c)
					wl = append(wl, &cidr)
				}

				transport = append(transport, &Transport{
					Name:         t.Name,
					FrontendPort: Port(t.SourcePort),
					BackendPort:  Port(destV1.Port),
					BackendPools: backendPools,
					Whitelist:    wl,
					HealthChecks: HealthChecks{
						Protocol: destV1.HealthChecks.Protocol,
						Path:     destV1.HealthChecks.Path,
						Code:     destV1.HealthChecks.Code,
					},
				})
			}
			vips[vipIdx] = &VIP{
				IP:        v.IP,
				Transport: transport,
			}
		}
		d.Spec[poolNameV0] = vips
	}
	return &d
}

type VIPV1 struct {
	IP        string `yaml:",omitempty"`
	Transport []*SourceV1
}

type SourceV1 struct {
	Name         string
	SourcePort   PortV1 `yaml:",omitempty"`
	Destinations []*DestinationV1
	Whitelist    []*CIDRV1
}

type DestinationV1 struct {
	HealthChecks HealthChecksV1
	Port         PortV1
	Pool         string
}

type CIDRV1 string

type PortV1 uint16

type HealthChecksV1 struct {
	Protocol string
	Path     string
	Code     uint16
}
