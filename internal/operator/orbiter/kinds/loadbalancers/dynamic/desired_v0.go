package dynamic

import (
	"github.com/caos/orbos/internal/tree"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   map[string][]*VIPV0
}

func v0tov1(v0 *DesiredV0) *DesiredV1 {
	d := DesiredV1{
		Spec: make(map[string][]*VIPV1, len(v0.Spec)),
	}
	for poolNameV0, poolV0 := range v0.Spec {
		vips := make([]*VIPV1, len(poolV0))
		for vipIdx, v := range poolV0 {

			var wl []*CIDRV1
			if vipIdx == 0 {
				for _, c := range v.Whitelist {
					cider := CIDRV1(*c)
					wl = append(wl, &cider)
				}
			}

			var transport []*SourceV1
			for _, t := range v.Transport {
				var dest []*DestinationV1
				for _, d := range t.Destinations {
					dest = append(dest, &DestinationV1{
						HealthChecks: HealthChecksV1{
							Protocol: d.HealthChecks.Protocol,
							Path:     d.HealthChecks.Path,
							Code:     d.HealthChecks.Code,
						},
						Port: PortV1(d.Port),
						Pool: d.Pool,
					})
				}

				transport = append(transport, &SourceV1{
					Name:         t.Name,
					SourcePort:   PortV1(t.SourcePort),
					Destinations: dest,
					Whitelist:    wl,
				})
			}
			vips[vipIdx] = &VIPV1{
				IP:        v.IP,
				Transport: transport,
			}
		}
		d.Spec[poolNameV0] = vips
	}
	return &d
}

type VIPV0 struct {
	IP        string
	Whitelist []*CIDRV0
	Transport []*SourceV0
}

type SourceV0 struct {
	Name         string
	SourcePort   PortV0 `yaml:",omitempty"`
	Destinations []*DestinationV0
}

type CIDRV0 string

type PortV0 uint16

type DestinationV0 struct {
	HealthChecks HealthChecksV0
	Port         PortV0
	Pool         string
}

type HealthChecksV0 struct {
	Protocol string
	Path     string
	Code     uint16
}
