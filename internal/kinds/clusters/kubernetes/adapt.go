package kubernetes

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

func AdaptFunc() orbiter.AdaptFunc {
	return func(k8sTree *orbiter.Tree, secrets *orbiter.Tree, nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent) (orbiter.EnsureFunc, error) {
		k8sKind := &KubernetesClusterV1{Common: k8sTree.Common}
		if err := k8sTree.Original.Decode(k8sKind); err != nil {
			panic(err)
		}
		k8sKind.Common.Version = "v1"
		k8sTree.Parsed = k8sKind
		return nil, nil
	}
}

type KubernetesClusterV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   struct {
		K8sVersion string
	}
	Deps map[string]*orbiter.Tree
}

type KubernetesClusterV1 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   struct {
		Versions struct {
			K8s string
		}
	}
	Deps map[string]*orbiter.Tree
}

func (k *KubernetesClusterV1) UnmarshalYAML(node *yaml.Node) error {
	switch k.Common.Version {
	case "v1":
		type latest KubernetesClusterV1
		l := latest{}
		if err := node.Decode(&l); err != nil {
			return err
		}
		k.Common = l.Common
		k.Spec = l.Spec
		k.Deps = l.Deps
		return nil
	case "v0":
		v0 := KubernetesClusterV0{}
		if err := node.Decode(&v0); err != nil {
			return err
		}
		k.Spec.Versions.K8s = v0.Spec.K8sVersion
		k.Deps = v0.Deps
		return nil
	}
	return errors.Errorf("Version %s for kind %s is not supported", k.Common.Version, k.Common.Kind)
}
