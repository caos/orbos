//go:generate goderive .

package dynamic

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

type Current struct {
	Common  *orbiter.Common `yaml:",inline"`
	Current struct {
		SourcePools map[string][]string
		Addresses   map[string]infra.Address
		Desire      func(pool string, svc core.ComputesService, nodeagents map[string]*common.NodeAgentSpec, notifyMaster string) error
	} `yaml:"-"`
}

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   map[string][]VIP
}

func (d *DesiredV0) Validate() error {

	ips := make([]string, 0)

	for pool, vips := range d.Spec {
		if len(vips) == 0 {
			return errors.Errorf("pool %s has no virtual ip configured", pool)
		}
		for _, vip := range vips {
			if err := vip.validate(); err != nil {
				return errors.Wrapf(err, "configuring vip for pool %s failed", pool)
			}
			ips = append(ips, vip.IP)
		}
	}

	if len(deriveUnique(ips)) != len(ips) {
		return errors.New("duplicate ips configured")
	}

	return nil
}

type VIP struct {
	IP        string
	Transport []Source
}

func (v *VIP) validate() error {
	if v.IP == "" {
		return errors.New("no virtual IP configured")
	}
	if len(v.Transport) == 0 {
		return errors.Errorf("vip %s has no transport configured", v.IP)
	}

	for _, source := range v.Transport {
		if err := source.validate(); err != nil {
			return errors.Wrapf(err, "configuring sources for vip %s failed", v.IP)
		}
	}

	withDestinations := len(deriveFilterSources(func(src Source) bool {
		return len(src.Destinations) > 0
	}, append([]Source(nil), v.Transport...)))

	if withDestinations != 0 && withDestinations != len(v.Transport) {
		return errors.Errorf("sources of vip %s must eighter all have configured destinations or none", v.IP)
	}

	return nil
}

type HealthChecks struct {
	Protocol string
	Path     string
	Code     uint8
}

func (h *HealthChecks) validate() error {
	if h.Protocol == "" {
		return errors.New("no protocol configured")
	}
	return nil
}

type Source struct {
	Name         string
	SourcePort   Port
	Destinations []Destination
}

func (s *Source) validate() error {
	if s.Name == "" {
		return errors.Errorf("source with port %d has no name", s.SourcePort)
	}

	if err := s.SourcePort.validate(); err != nil {
		return errors.Wrapf(err, "configuring port for source %s failed", s.Name)
	}

	for _, dest := range s.Destinations {
		if err := dest.validate(); err != nil {
			return errors.Wrapf(err, "configuring destinations for source %s failed", s.Name)
		}
	}

	return nil
}

type Destination struct {
	HealthChecks HealthChecks
	Port         Port
	Pool         string
}

func (d *Destination) validate() error {

	if d.Pool == "" {
		return errors.New("destination with port %d has no pool configured")
	}

	if err := d.Port.validate(); err != nil {
		return errors.Wrapf(err, "configuring port for destination with pool %s failed", d.Pool)
	}

	return errors.Wrapf(d.HealthChecks.validate(), "configuring healthchecks for destination with pool %s failed", d.Pool)
}

type Port uint16

func (p Port) validate() error {
	if p == 0 {
		return errors.Errorf("port %d is not allowed", p)
	}
	return nil
}
