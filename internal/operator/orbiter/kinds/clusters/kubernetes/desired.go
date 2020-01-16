package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
)

type DesiredV0 struct {
	Common orbiter.Common `yaml:",inline"`
	Deps   map[string]*orbiter.Tree
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
