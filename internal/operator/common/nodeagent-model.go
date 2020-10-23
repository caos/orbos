//go:generate goderive -autoname -dedup .

package common

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type NodeAgentSpec struct {
	ChangesAllowed bool
	//	RebootEnabled  bool
	Software       *Software
	Firewall       *Firewall
	RebootRequired time.Time
}

type NodeAgentCurrent struct {
	NodeIsReady bool `mapstructure:"ready" yaml:"ready"`
	Software    Software
	Open        []*ZoneDesc
	Commit      string
	Booted      time.Time
}

type Software struct {
	Swap             Package `yaml:",omitempty"`
	Kubelet          Package `yaml:",omitempty"`
	Kubeadm          Package `yaml:",omitempty"`
	Kubectl          Package `yaml:",omitempty"`
	Containerruntime Package `yaml:",omitempty"`
	KeepaliveD       Package `yaml:",omitempty"`
	Nginx            Package `yaml:",omitempty"`
	SSHD             Package `yaml:",omitempty"`
	Hostname         Package `yaml:",omitempty"`
	Sysctl           Package `yaml:",omitempty"`
	Health           Package `yaml:",omitempty"`
}

func (s *Software) Merge(sw Software) {

	zeroPkg := Package{}

	if !sw.Containerruntime.Equals(zeroPkg) {
		s.Containerruntime = sw.Containerruntime
	}

	if !sw.KeepaliveD.Equals(zeroPkg) {
		s.KeepaliveD = sw.KeepaliveD
	}

	if !sw.Nginx.Equals(zeroPkg) {
		s.Nginx = sw.Nginx
	}

	if !sw.Kubeadm.Equals(zeroPkg) {
		s.Kubeadm = sw.Kubeadm
	}

	if !sw.Kubelet.Equals(zeroPkg) {
		s.Kubelet = sw.Kubelet
	}

	if !sw.Kubectl.Equals(zeroPkg) {
		s.Kubectl = sw.Kubectl
	}

	if !sw.Swap.Equals(zeroPkg) {
		s.Swap = sw.Swap
	}

	if !sw.SSHD.Equals(zeroPkg) {
		s.SSHD = sw.SSHD
	}

	if !sw.Hostname.Equals(zeroPkg) {
		s.Hostname = sw.Hostname
	}

	if !sw.Sysctl.Equals(zeroPkg) && s.Sysctl.Config == nil {
		s.Sysctl.Config = make(map[string]string)
	}
	for key, value := range sw.Sysctl.Config {
		s.Sysctl.Config[key] = value
	}
	if !sw.Health.Equals(zeroPkg) && s.Health.Config == nil {
		s.Health.Config = make(map[string]string)
	}
	for key, value := range sw.Health.Config {
		s.Health.Config[key] = value
	}
}

type Package struct {
	Version string            `yaml:",omitempty"`
	Config  map[string]string `yaml:",omitempty"`
}

var prune = regexp.MustCompile("[^a-zA-Z0-9]+")

func configEquals(this, that map[string]string) bool {
	if this == nil || that == nil {
		return this == nil && that == nil
	}
	if len(this) != len(that) {
		return false
	}
	for k, v := range this {
		thatv, ok := that[k]
		if !ok {
			return false
		}

		if prune.ReplaceAllString(v, "") != prune.ReplaceAllString(thatv, "") {
			return false
		}
	}
	return true
}

func (p Package) Equals(other Package) bool {
	return PackageEquals(p, other)
}
func PackageEquals(this, that Package) bool {
	equals := this.Version == that.Version &&
		configEquals(this.Config, that.Config)
	return equals
}

type Firewall struct {
	mux   sync.Mutex       `yaml:"-"`
	Zones map[string]*Zone `yaml:",inline"`
}

type Zone struct {
	Interfaces []string
	FW         map[string]*Allowed
	Services   map[string]*Service
}

type Service struct {
	Description string
	Ports       []*Allowed
}

func ToFirewall(zone string, fw map[string]*Allowed) Firewall {
	return Firewall{Zones: map[string]*Zone{zone: {Interfaces: []string{}, FW: fw, Services: map[string]*Service{}}}}
}

func (f *Firewall) Merge(fw Firewall) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.Zones == nil {
		f.Zones = make(map[string]*Zone, 0)
	}

	if fw.Zones == nil {
		return
	}

	for name, zone := range fw.Zones {
		if zone == nil {
			continue
		}

		current, ok := f.Zones[name]
		if !ok || current == nil {
			current = &Zone{}
		}

		if zone.FW != nil {
			if current.FW == nil {
				current.FW = make(map[string]*Allowed)
			}

			for key, value := range zone.FW {
				current.FW[key] = value
			}
		}

		if zone.Interfaces != nil {
			if current.Interfaces == nil {
				current.Interfaces = []string{}
			}
			for _, i := range zone.Interfaces {
				found := false
				for _, i2 := range current.Interfaces {
					if i == i2 {
						found = true
					}
				}
				if !found {
					current.Interfaces = append(current.Interfaces, i)
				}
			}

		}
		if zone.Services != nil {
			if current.Services == nil {
				current.Services = make(map[string]*Service, 0)
			}

			for key, value := range zone.Services {
				current.Services[key] = value
			}
		}

		f.Zones[name] = current
	}
}

func (f *Firewall) AllZones() []*ZoneDesc {
	zones := make([]*ZoneDesc, 0)
	if f.Zones == nil {
		return zones
	}

	for name, zone := range f.Zones {
		if zone != nil {
			zones = append(zones, &ZoneDesc{
				Name:       name,
				Interfaces: zone.Interfaces,
				Services:   []*Service{},
				FW:         f.Ports(name),
			})
		}
	}
	return zones
}

func (f *Firewall) Ports(zoneName string) Ports {
	ports := make([]*Allowed, 0)
	if f.Zones == nil {
		return ports
	}
	for name, zone := range f.Zones {
		if name == zoneName && zone != nil && zone.FW != nil {
			for _, value := range zone.FW {
				ports = append(ports, value)
			}
		}
	}
	return ports
}

type ZoneDesc struct {
	Name       string
	Interfaces []string
	FW         []*Allowed
	Services   []*Service
}

type Ports []*Allowed

func (p Ports) String() string {
	strs := make([]string, len(p))
	for idx, port := range p {
		strs[idx] = fmt.Sprintf("%s/%s", port.Port, port.Protocol)
	}
	return strings.Join(strs, " ")
}

type Allowed struct {
	Port     string
	Protocol string
}

func (f Firewall) Contains(other Firewall) bool {
	if other.Zones == nil {
		return true
	}

	for name, zone := range other.Zones {
		current, ok := f.Zones[name]
		if !ok || current == nil {
			return false
		}

		if current.FW == nil {
			return false
		}
		for name, port := range zone.FW {

			found, ok := current.FW[name]
			if !ok {
				return false
			}
			if !deriveEqualPort(*port, *found) {
				return false
			}
		}
	}
	return true
}

func (f Firewall) IsContainedIn(zones []*ZoneDesc) bool {
	if zones == nil {
		return false
	}
	if f.Zones == nil {
		return true
	}

	for _, currentZone := range zones {
		found := false

		for name, zone := range f.Zones {
			if currentZone.Name == name {
				if (currentZone.FW == nil || len(currentZone.FW) == 0) && (zone.FW != nil || len(zone.FW) > 0) {
					continue
				}
				for _, currentPort := range currentZone.FW {
					for _, fwPort := range zone.FW {
						if deriveEqualPort(*currentPort, *fwPort) {
							found = true
						}
					}
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type NodeAgentsCurrentKind struct {
	Kind    string
	Version string
	Current CurrentNodeAgents
}

type CurrentNodeAgents struct {
	// NA is exported for yaml (de)serialization and not intended to be accessed by any other code outside this package
	NA  map[string]*NodeAgentCurrent `yaml:",inline"`
	mux sync.Mutex                   `yaml:"-"`
}

func (n *CurrentNodeAgents) Set(id string, na *NodeAgentCurrent) {
	n.mux.Lock()
	defer n.mux.Unlock()
	if n.NA == nil {
		n.NA = make(map[string]*NodeAgentCurrent)
	}
	n.NA[id] = na
}

func (n *CurrentNodeAgents) Get(id string) (*NodeAgentCurrent, bool) {
	n.mux.Lock()
	defer n.mux.Unlock()

	if n.NA == nil {
		n.NA = make(map[string]*NodeAgentCurrent)
	}

	na, ok := n.NA[id]
	if !ok {
		na = &NodeAgentCurrent{
			Open: make([]*ZoneDesc, 0),
		}
		n.NA[id] = na
	}
	return na, ok

}

type NodeAgentsSpec struct {
	Commit     string
	NodeAgents DesiredNodeAgents
}

type DesiredNodeAgents struct {
	// NA is exported for yaml (de)serialization and not intended to be accessed by any other code outside this package
	NA  map[string]*NodeAgentSpec `yaml:",inline"`
	mux sync.Mutex                `yaml:"-"`
}

func (n *DesiredNodeAgents) Delete(id string) {
	n.mux.Lock()
	defer n.mux.Unlock()
	delete(n.NA, id)
}

func (n *DesiredNodeAgents) Get(id string) (*NodeAgentSpec, bool) {
	n.mux.Lock()
	defer n.mux.Unlock()

	if n.NA == nil {
		n.NA = make(map[string]*NodeAgentSpec)
	}

	na, ok := n.NA[id]
	if !ok {
		na = &NodeAgentSpec{
			Software: &Software{},
			Firewall: &Firewall{
				Zones: map[string]*Zone{
					"internal": {
						Interfaces: []string{},
						FW:         map[string]*Allowed{},
						Services:   map[string]*Service{},
					}, "external": {
						Interfaces: []string{},
						FW:         map[string]*Allowed{},
						Services:   map[string]*Service{},
					},
				},
			},
		}
		n.NA[id] = na
	}
	return na, ok
}

type NodeAgentsDesiredKind struct {
	Kind    string
	Version string
	Spec    NodeAgentsSpec `yaml:",omitempty"`
}
