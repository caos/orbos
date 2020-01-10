//go:generate goderive -autoname -dedup .

package operator

import (
	"regexp"
	"sync"

	"github.com/mitchellh/mapstructure"

	"github.com/caos/orbiter/logging"
)

var nodeagentBytesGZIPBase64 string

type NodeAgentCurrent struct {
	NodeIsReady bool `mapstructure:"ready" yaml:"ready"`
	Software    Software
	Open        Firewall
	Commit      string
	changer     *changer
}

type Software struct {
	Swap             Package `yaml:",omitempty"`
	Kubelet          Package `yaml:",omitempty"`
	Kubeadm          Package `yaml:",omitempty"`
	Kubectl          Package `yaml:",omitempty"`
	Containerruntime Package `yaml:",omitempty"`
	KeepaliveD       Package `yaml:",omitempty"`
	Nginx            Package `yaml:",omitempty"`
	Hostname         Package `yaml:",omitempty"`
}

type Package struct {
	Version string            `yaml:",omitempty"`
	Config  map[string]string `yaml:",omitempty"`
}

var prune = regexp.MustCompile("[^a-zA-Z0-9]+")

func configEquals(this, that map[string]string) bool {
	if this == nil || that == nil {
		equal := this == nil && that == nil
		return equal
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

func packageEquals(this, that Package) bool {
	return this.Version == that.Version &&
		configEquals(this.Config, that.Config)
}

func contains(this, that Package) bool {
	return that.Version == "" && that.Config == nil || packageEquals(this, that)
}

func (p Package) Equals(other Package) bool {
	return packageEquals(p, other)
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

type Firewall map[string]Allowed

func (f Firewall) Contains(other Firewall) bool {
	for name, port := range other {
		found, ok := f[name]
		if !ok {
			return false
		}
		if !deriveEqualPort(port, found) {
			return false
		}
	}
	return true
}

type Allowed struct {
	Port     string
	Protocol string
}

type NodeAgentSpec struct {
	Commit         string
	ChangesAllowed bool
	//	RebootEnabled  bool
	Software *Software
	Firewall Firewall
}

func (n *NodeAgentCurrent) AllowChanges() {
	n.changer.desire(func(spec *NodeAgentSpec) {
		spec.ChangesAllowed = true
	})
}

func (n *NodeAgentCurrent) DesireFirewall(fw Firewall) {
	n.changer.desire(func(spec *NodeAgentSpec) {
		if spec.Firewall == nil {
			spec.Firewall = make(map[string]Allowed)
		}
		for key, value := range fw {
			spec.Firewall[key] = value
		}
	})
}

func (n *NodeAgentCurrent) DesireSoftware(sw Software) {
	n.changer.desire(func(spec *NodeAgentSpec) {
		if spec.Software == nil {
			spec.Software = &Software{}
		}

		zeroPkg := Package{}

		if !sw.Containerruntime.Equals(zeroPkg) {
			spec.Software.Containerruntime = sw.Containerruntime
		}

		if !sw.KeepaliveD.Equals(zeroPkg) {
			spec.Software.KeepaliveD = sw.KeepaliveD
		}

		if !sw.Nginx.Equals(zeroPkg) {
			spec.Software.Nginx = sw.Nginx
		}

		if !sw.Kubeadm.Equals(zeroPkg) {
			spec.Software.Kubeadm = sw.Kubeadm
		}

		if !sw.Kubelet.Equals(zeroPkg) {
			spec.Software.Kubelet = sw.Kubelet
		}

		if !sw.Kubectl.Equals(zeroPkg) {
			spec.Software.Kubectl = sw.Kubectl
		}

		if !sw.Swap.Equals(zeroPkg) {
			spec.Software.Swap = sw.Swap
		}

		if !sw.Hostname.Equals(zeroPkg) {
			spec.Software.Hostname = sw.Hostname
		}
	})
}

type changer struct {
	path    []string
	kind    map[string]interface{}
	changes chan<- *nodeAgentChange
}

func (c *changer) desire(mutate func(*NodeAgentSpec)) {
	newSpec := &NodeAgentSpec{}
	mapstructure.Decode(c.kind["spec"], newSpec)
	mutate(newSpec)
	c.kind["spec"] = newSpec
	c.changes <- &nodeAgentChange{
		path: c.path,
		spec: newSpec,
	}
}

type NodeAgentKind struct {
	Kind    string
	Version string
	Spec    interface{}
	Current *NodeAgentCurrent `yaml:",omitempty"`
}

type muxMap struct {
	mux  sync.Mutex
	data map[string]interface{}
}

func newNodeAgentCurrent(logger logging.Logger, path []string, containingKind map[string]interface{}, changes chan<- *nodeAgentChange) *NodeAgentCurrent {

	naKind, err := drillIn(logger.WithFields(map[string]interface{}{
		"purpose": "find node agent",
		"config":  "current",
	}), containingKind, append([]string{"current"}, path...), true)
	if err != nil {
		panic(err)
	}

	kind := &NodeAgentKind{}
	mapstructure.Decode(naKind, kind)
	if kind.Current == nil {
		kind.Current = &NodeAgentCurrent{}
	}
	kind.Current.changer = &changer{path, naKind, changes}
	return kind.Current
}
