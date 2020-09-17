package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel/configuration"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configuration for the zitadel deployment
	Spec *Spec
	//Configuration for the used database for zitadel
	Database *tree.Tree `yaml:"database"`
	//Configuration for the networking for zitadel
	Networking *tree.Tree `yaml:"networking"`
}

type Spec struct {
	//Verbose flag to set debug-level to debug
	Verbose bool
	//Number of replicas for zitadel
	ReplicaCount int `yaml:"replicaCount,omitempty"`
	//Configuration for zitadel
	Configuration *configuration.Configuration `yaml:"configuration"`
	//Node-selector to let zitadel only on specific nodes
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	//Tolerations on node-taints for zitadel
	Tolerations []corev1.Toleration `yaml:"tolerations,omitempty"`
	//Affinity for zitadel
	Affinity *k8s.Affinity `yaml:"affinity,omitempty"`
	//Definition for resource limits and requests for zitadel
	Resources *k8s.Resources `yaml:"resources,omitempty"`
}

func (s *Spec) IsZero() bool {
	if (s.Configuration == nil || s.Configuration.IsZero()) &&
		!s.Verbose &&
		s.ReplicaCount == 0 &&
		s.NodeSelector == nil &&
		s.Tolerations == nil &&
		s.Affinity == nil &&
		s.Resources == nil {
		return true
	}
	return false
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   &Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
