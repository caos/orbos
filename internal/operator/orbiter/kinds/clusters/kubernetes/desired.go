package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
)

type DesiredV0 struct {
	Common orbiter.Common `yaml:",inline"`
	Spec   struct {
		Verbose  bool
		Versions struct {
			Kubernetes string
			Orbiter    string
			Boom       string
		}
		Networking struct {
			DNSDomain   string
			Network     string
			ServiceCidr string
			PodCidr     string
		}
		ControlPlane Pool
		Workers      map[string]*Pool
	}
}

func (d *DesiredV0) validate() error {

	if d.Spec.ControlPlane.Nodes != 1 && d.Spec.ControlPlane.Nodes != 3 && d.Spec.ControlPlane.Nodes != 5 {
		return errors.Errorf("Controlplane nodes can only be scaled to 1, 3 or 5 but desired are %d", d.Spec.ControlPlane.Nodes)
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
	switch d.Common.Version {
	case "v1":
		type latest DesiredV1
		l := latest{}
		if err := node.Decode(&l); err != nil {
			return err
		}
		d.Common = l.Common
		d.Spec = l.Spec
		d.Deps = l.Deps
		return nil
	case "v0":
		v0 := DesiredV0{}
		if err := node.Decode(&v0); err != nil {
			return err
		}
		d.Spec.Versions.Kubernetes = v0.Spec.Kubernetes
		d.Deps = v0.Deps
		return nil
	}
	return errors.Errorf("Version %s for kind %s is not supported", d.Common.Version, d.Common.Kind)
}
*/
