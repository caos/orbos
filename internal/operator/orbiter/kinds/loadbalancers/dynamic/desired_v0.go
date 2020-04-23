package dynamic

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
	"gopkg.in/yaml.v3"
)

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   map[string][]*VIPV0
}

func v0tov1(node *yaml.Node, d *Desired) error {
	v0 := DesiredV0{}
	if err := node.Decode(&v0); err != nil {
		return err
	}
	d.Spec = make(map[string][]*VIP, len(v0.Spec))
	for poolNameV0, poolV0 := range v0.Spec {
		vips := make([]*VIP, len(poolV0))
		for vipIdx, vipV0 := range poolV0 {
			for _, transportV0 := range vipV0.Transport {
				if transportV0.Whitelist == nil {
					transportV0.Whitelist = vipV0.Whitelist
				}
			}
			vips[vipIdx] = &VIP{
				IP:        vipV0.IP,
				Transport: vipV0.Transport,
			}
		}
		d.Spec[poolNameV0] = vips
	}
	return nil
}

type VIPV0 struct {
	IP        string
	Whitelist []*orbiter.CIDR
	Transport []*Source
}
