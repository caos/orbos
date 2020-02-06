package kubernetes

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
)

var ipPartRegex = `([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])`

var ipRegex = fmt.Sprintf(`%s\.%s\.%s\.%s`, ipPartRegex, ipPartRegex, ipPartRegex, ipPartRegex)

var cidrRegex = fmt.Sprintf(`%s/([1-2][0-9]|3[0-2]|[0-9])`, ipRegex)

var cidrComp = regexp.MustCompile(fmt.Sprintf(`^(%s)$`, cidrRegex))

type cidr string

/*
type DesiredV1 struct {
	Common orbiter.Common `yaml:",inline"`
	Spec   struct {
		Verbose           bool
		KubernetesVersion string
		Versions          struct {
			Orbiter string
			Boom    string
		}
		Networking struct {
			DNSDomain   string
			Network     string
			ServiceCidr orbiter.CIDR
			PodCidr     orbiter.CIDR
		}
		ControlPlane Pool
		Workers      []*Pool
	}
}
*/
type DesiredV0 struct {
	Common orbiter.Common `yaml:",inline"`
	Spec   Spec
}

type Spec struct {
	ControlPlane Pool
	Kubeconfig   *orbiter.Secret `yaml:",omitempty"`
	Networking   struct {
		DNSDomain   string
		Network     string
		ServiceCidr orbiter.CIDR
		PodCidr     orbiter.CIDR
	}
	Verbose  bool
	Versions struct {
		Kubernetes string
		Orbiter    string
		Boom       string
	}
	Workers []*Pool
}

func (d *DesiredV0) validate() error {

	if d.Spec.ControlPlane.Nodes != 1 && d.Spec.ControlPlane.Nodes != 3 && d.Spec.ControlPlane.Nodes != 5 {
		return errors.Errorf("Controlplane nodes can only be scaled to 1, 3 or 5 but desired are %d", d.Spec.ControlPlane.Nodes)
	}

	if k8s.ParseString(d.Spec.Versions.Kubernetes) == k8s.Unknown {
		return errors.Errorf("Unknown kubernetes version %s", d.Spec.Versions.Kubernetes)
	}

	if d.Spec.Networking.Network != "cilium" && d.Spec.Networking.Network != "calico" {
		return errors.Errorf("Network must eighter be calico or cilium, but got %s", d.Spec.Networking.Network)
	}

	if err := d.Spec.Networking.ServiceCidr.Validate(); err != nil {
		return err
	}

	if err := d.Spec.Networking.PodCidr.Validate(); err != nil {
		return err
	}

	seenPools := map[string][]string{
		d.Spec.ControlPlane.Provider: []string{d.Spec.ControlPlane.Pool},
	}

	for _, worker := range d.Spec.Workers {
		pools, ok := seenPools[worker.Provider]
		if !ok {
			seenPools[worker.Provider] = []string{worker.Pool}
			continue
		}
		for _, seenPool := range pools {
			if seenPool == worker.Pool {
				return errors.Errorf("Pool %s from provider %s is used multiple times", worker.Pool, worker.Provider)
			}
		}
		seenPools[worker.Provider] = append(pools, worker.Pool)
	}

	return nil
}

type Pool struct {
	UpdatesDisabled bool
	Provider        string
	Nodes           int
	Pool            string
}

/*
// UnmarshalYAML migrates desired states from v0 to v1:
func (d *DesiredV1) UnmarshalYAML(node *yaml.Node) error {
	defer func() {
		d.Common.Version = "v1"
	}()
	switch d.Common.Version {
	case "v1":
		type latest DesiredV1
		l := latest{}
		if err := node.Decode(&l); err != nil {
			return err
		}
		d.Common = l.Common
		d.Spec = l.Spec
		return nil
	case "v0":
		v0 := DesiredV0{}
		if err := node.Decode(&v0); err != nil {
			return err
		}
		d.Spec.Verbose = v0.Spec.Verbose
		d.Spec.KubernetesVersion = v0.Spec.Versions.Kubernetes
		d.Spec.Versions.Orbiter = v0.Spec.Versions.Orbiter
		d.Spec.Versions.Boom = v0.Spec.Versions.Boom
		d.Spec.Networking = v0.Spec.Networking
		d.Spec.ControlPlane = v0.Spec.ControlPlane
		d.Spec.Workers = v0.Spec.Workers
		return nil
	}
	return errors.Errorf("Version %s for kind %s is not supported", d.Common.Version, d.Common.Kind)
}
*/
