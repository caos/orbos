//go:generate goderive -autoname -dedup .

package common

import (
	"regexp"
)

type NodeAgentSpec struct {
	ChangesAllowed bool
	//	RebootEnabled  bool
	Software *Software
	Firewall *Firewall
}

type NodeAgentCurrent struct {
	NodeIsReady bool `mapstructure:"ready" yaml:"ready"`
	Software    Software
	Open        []*Allowed
	Commit      string
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
	return packageEquals(p, other)
}
func packageEquals(this, that Package) bool {
	equals := this.Version == that.Version &&
		configEquals(this.Config, that.Config)
	return equals
}

func (this *Software) Contains(that Software) bool {
	return contains(this.Swap, that.Swap) &&
		contains(this.Kubelet, that.Kubelet) &&
		contains(this.Kubeadm, that.Kubeadm) &&
		contains(this.Kubectl, that.Kubectl) &&
		contains(this.Containerruntime, that.Containerruntime) &&
		contains(this.KeepaliveD, that.KeepaliveD) &&
		contains(this.Nginx, that.Nginx) &&
		contains(this.Hostname, that.Hostname)
}

func contains(this, that Package) bool {
	return that.Version == "" && that.Config == nil || packageEquals(this, that)
}

func (this *Software) Defines(that Software) bool {
	return defines(this.Swap, that.Swap) &&
		defines(this.Kubelet, that.Kubelet) &&
		defines(this.Kubeadm, that.Kubeadm) &&
		defines(this.Kubectl, that.Kubectl) &&
		defines(this.Containerruntime, that.Containerruntime) &&
		defines(this.KeepaliveD, that.KeepaliveD) &&
		defines(this.Nginx, that.Nginx) &&
		defines(this.Hostname, that.Hostname)
}

func defines(this, that Package) bool {
	zeroPkg := Package{}
	defines := packageEquals(that, zeroPkg) || !packageEquals(this, zeroPkg)
	return defines
}

type Firewall map[string]*Allowed

func (f *Firewall) Merge(fw Firewall) {
	for key, value := range fw {
		m := *f
		m[key] = value
	}
}

func (f *Firewall) Ports() []*Allowed {
	ports := make([]*Allowed, 0)
	for _, value := range *f {
		ports = append(ports, value)
	}
	return ports
}

type Allowed struct {
	Port     string
	Protocol string
}

func (f Firewall) Contains(other Firewall) bool {
	for name, port := range other {
		found, ok := f[name]
		if !ok {
			return false
		}
		if !deriveEqualPort(*port, *found) {
			return false
		}
	}
	return true
}

func (f Firewall) IsContainedIn(ports []*Allowed) bool {
	if len(f) > len(ports) {
		return false
	}
checks:
	for _, fwPort := range f {
		for _, port := range ports {
			if deriveEqualPort(*port, *fwPort) {
				continue checks
			}
		}
		return false
	}
	return true
}

type NodeAgentsCurrentKind struct {
	Kind    string
	Version string
	Current map[string]*NodeAgentCurrent `yaml:",omitempty"`
}

type NodeAgentsSpec struct {
	Commit     string
	NodeAgents map[string]*NodeAgentSpec
}

type NodeAgentsDesiredKind struct {
	Kind    string
	Version string
	Spec    NodeAgentsSpec `yaml:",omitempty"`
}
