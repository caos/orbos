//go:generate goderive .

package dynamic

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/operator/orbiter"
)

type Desired struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   map[string][]*VIP
}

func (d *Desired) UnmarshalYAML(node *yaml.Node) (err error) {
	defer func() {
		d.Common.Version = "v1"
	}()
	switch d.Common.Version {
	case "v1":
		type latest Desired
		l := latest{}
		if err := node.Decode(&l); err != nil {
			return err
		}
		d.Spec = l.Spec
		return nil
	case "v0":
		return v0tov1(node, d)
	}
	return errors.Errorf("Version %s for kind %s is not supported", d.Common.Version, d.Common.Kind)
}

func (d *Desired) Validate() error {

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
	Transport []*Source
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

	withDestinations := len(deriveFilterSources(func(src *Source) bool {
		return len(src.Destinations) > 0
	}, append([]*Source(nil), v.Transport...)))

	if withDestinations != 0 && withDestinations != len(v.Transport) {
		return errors.Errorf("sources of vip %s must eighter all have configured destinations or none", v.IP)
	}

	return nil
}

type HealthChecks struct {
	Protocol string
	Path     string
	Code     uint16
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
	Destinations []*Destination
	Whitelist    []*orbiter.CIDR
}

func (s *Source) validate() (err error) {

	defer func() {
		err = errors.Wrapf(err, "source %s is invalid", s.Name)
	}()

	if s.Name == "" {
		return errors.Errorf("source with port %d has no name", s.SourcePort)
	}

	for _, cidr := range s.Whitelist {
		if err := cidr.Validate(); err != nil {
			return err
		}
	}

	if err := s.SourcePort.validate(); err != nil {
		return err
	}

	for _, dest := range s.Destinations {
		if err := dest.validate(); err != nil {
			return err
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
