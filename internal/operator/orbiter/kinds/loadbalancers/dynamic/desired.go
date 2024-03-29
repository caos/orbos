//go:generate goderive .

package dynamic

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/tree"
)

type Desired struct {
	Common *tree.Common `yaml:",inline"`
	Spec   map[string][]*VIP
}

func (d *Desired) UnmarshalYAML(node *yaml.Node) (err error) {
	defer func() {
		err = mntr.ToUserError(err)
		d.Common.OverwriteVersion("v2")
	}()
	switch d.Common.Version() {
	case "v2":
		type latest Desired
		l := latest{}
		if err := node.Decode(&l); err != nil {
			return err
		}
		d.Spec = l.Spec
		return nil
	case "v1":
		v1 := &DesiredV1{}
		if err := node.Decode(v1); err != nil {
			return err
		}

		d.Spec = v1tov2(v1).Spec
		return nil
	case "v0":
		v0 := &DesiredV0{}
		if err := node.Decode(v0); err != nil {
			return err
		}

		d.Spec = v1tov2(v0tov1(v0)).Spec
		return nil
	}
	return fmt.Errorf("version %s for kind %s is not supported", d.Common.Version, d.Common.Kind)
}

func (d *Desired) Validate() (err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	ips := make([]string, 0)

	for pool, vips := range d.Spec {
		if len(vips) == 0 {
			return fmt.Errorf("pool %s has no virtual ip configured", pool)
		}
		for _, vip := range vips {
			if err := vip.validate(); err != nil {
				return fmt.Errorf("configuring vip for pool %s failed: %w", pool, err)
			}
			if vip != nil && vip.IP != "" {
				ips = append(ips, vip.IP)
			}
		}
	}

	if len(deriveUnique(ips)) != len(ips) {
		return errors.New("duplicate ips configured")
	}

	return nil
}

type VIP struct {
	IP        string `yaml:",omitempty"`
	Transport []*Transport
}

func (v *VIP) validate() (err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	if len(v.Transport) == 0 {
		return fmt.Errorf("vip %s has no transport configured", v.IP)
	}

	for _, source := range v.Transport {
		if err := source.validate(); err != nil {
			return fmt.Errorf("configuring sources for vip %s failed: %w", v.IP, err)
		}
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
		return mntr.ToUserError(errors.New("no protocol configured"))
	}
	return nil
}

func (s *Transport) validate() (err error) {

	defer func() {
		if err != nil {
			mntr.ToUserError(fmt.Errorf("source %s is invalid: %w", s.Name, err))
		}
	}()

	if s.Name == "" {
		return fmt.Errorf("source with port %d has no name", s.FrontendPort)
	}

	if err := s.FrontendPort.validate(); err != nil {
		return fmt.Errorf("configuring frontend port failed: %w", err)
	}

	if err := s.BackendPort.validate(); err != nil {
		return fmt.Errorf("configuring backend port failed: %w", err)
	}

	if s.FrontendPort == s.BackendPort {
		return errors.New("frontend port and backend port must not be equal")
	}

	for _, cidr := range s.Whitelist {
		if err := cidr.Validate(); err != nil {
			return err
		}
	}

	if len(s.BackendPools) < 1 {
		return errors.New("at least one target pool is needed")
	}

	if err := s.HealthChecks.validate(); err != nil {
		return fmt.Errorf("configuring health checks failed: %w", err)
	}

	return nil
}

type Transport struct {
	Name         string
	FrontendPort Port
	BackendPort  Port
	BackendPools []string
	Whitelist    []*orbiter.CIDR
	//	DownstreamProxies []*orbiter.IPAddress
	HealthChecks  HealthChecks
	ProxyProtocol *bool
}

type Port uint16

func (p Port) validate() error {
	if p == 0 {
		return mntr.ToUserError(fmt.Errorf("port %d is not allowed", p))
	}
	return nil
}
