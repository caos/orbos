package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel/configuration"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type DesiredV0 struct {
	Common     *tree.Common `yaml:",inline"`
	Spec       *Spec
	Database   *tree.Tree `yaml:"database"`
	Networking *tree.Tree `yaml:"networking"`
}

type Spec struct {
	Verbose       bool
	ReplicaCount  int                          `yaml:"replicaCount,omitempty"`
	Configuration *configuration.Configuration `yaml:"configuration"`
	NodeSelector  map[string]string            `yaml:"nodeSelector,omitempty"`
	Tolerations   []corev1.Toleration          `yaml:"tolerations,omitempty"`
	Affinity      *k8s.Affinity                `yaml:"affinity,omitempty"`
	Resources     *k8s.Resources               `yaml:"resources,omitempty"`
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
