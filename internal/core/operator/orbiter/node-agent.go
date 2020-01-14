//go:generate goderive -autoname -dedup .

package orbiter

import (
	"regexp"
)

type NodeAgentSpec struct {
	ChangesAllowed bool
	//	RebootEnabled  bool
	Software *Software
	Firewall Firewall
}

type NodeAgentCurrent struct {
	NodeIsReady bool `mapstructure:"ready" yaml:"ready"`
	Software    Software
	Open        Firewall
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
		if !deriveEqualPort(port, found) {
			return false
		}
	}
	return true
}

/*
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
		} else if !n.Software.Containerruntime.Equals(zeroPkg) {
			spec.Software.Containerruntime = n.Software.Containerruntime
		}

		if !sw.KeepaliveD.Equals(zeroPkg) {
			spec.Software.KeepaliveD = sw.KeepaliveD
		} else if !n.Software.KeepaliveD.Equals(zeroPkg) {
			spec.Software.KeepaliveD = n.Software.KeepaliveD
		}

		if !sw.Nginx.Equals(zeroPkg) {
			spec.Software.Nginx = sw.Nginx
		} else if !n.Software.Nginx.Equals(zeroPkg) {
			spec.Software.Nginx = n.Software.Nginx
		}

		if !sw.Kubeadm.Equals(zeroPkg) {
			spec.Software.Kubeadm = sw.Kubeadm
		} else if !n.Software.Kubeadm.Equals(zeroPkg) {
			spec.Software.Kubeadm = n.Software.Kubeadm
		}

		if !sw.Kubelet.Equals(zeroPkg) {
			spec.Software.Kubelet = sw.Kubelet
		} else if !n.Software.Kubelet.Equals(zeroPkg) {
			spec.Software.Kubelet = n.Software.Kubelet
		}

		if !sw.Kubectl.Equals(zeroPkg) {
			spec.Software.Kubectl = sw.Kubectl
		} else if !n.Software.Kubectl.Equals(zeroPkg) {
			spec.Software.Kubectl = n.Software.Kubectl
		}

		if !sw.Swap.Equals(zeroPkg) {
			spec.Software.Swap = sw.Swap
		} else if !n.Software.Swap.Equals(zeroPkg) {
			spec.Software.Swap = n.Software.Swap
		}

		if !sw.Hostname.Equals(zeroPkg) {
			spec.Software.Hostname = sw.Hostname
		} else if !n.Software.Hostname.Equals(zeroPkg) {
			spec.Software.Hostname = n.Software.Hostname
		}
	})
}

type changer struct {
	id      string
	changes chan<- *nodeAgentChange
}

func (c *changer) desire(mutate func(*NodeAgentSpec)) {
	c.changes <- &nodeAgentChange{
		id:     c.id,
		mutate: mutate,
	}
}
*/
type NodeAgentsCurrentKind struct {
	Common  `yaml:",inline"`
	Current map[string]*NodeAgentCurrent `yaml:",omitempty"`
}

type NodeAgentsSpec struct {
	Commit     string
	NodeAgents map[string]*NodeAgentSpec
}

type NodeAgentsDesiredKind struct {
	Common `yaml:",inline"`
	Spec   NodeAgentsSpec `yaml:",omitempty"`
}

/*
type muxMap struct {
	mux  sync.Mutex
	data map[string]interface{}
}

func newNodeAgentCurrentFunc(
	logger logging.Logger,
	current []byte) func(id string, changes chan<- *nodeAgentChange) *NodeAgentCurrent {

	nodeagents := NodeAgentsCurrentKind{}
	yaml.Unmarshal(current, &nodeagents)

	return func(id string, changes chan<- *nodeAgentChange) *NodeAgentCurrent {

		curr, ok := nodeagents.Current[id]
		if !ok {
			curr = &NodeAgentCurrent{}
		}

		curr.changer = &changer{id, changes}
		return curr
	}
}
*/
