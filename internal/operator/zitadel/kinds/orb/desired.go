package orb

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configuration for zitadel-operator
	Spec Spec
	//Configuration for the IAM
	IAM *tree.Tree `yaml:"iam"`
}
type Spec struct {
	//Verbose flag to set debug-level to debug
	Verbose bool
	//Node-selector to let zitadel-operator only on specific nodes
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	//Tolerations on node-taints for zitadel-operator
	Tolerations []corev1.Toleration `yaml:"tolerations,omitempty"`
	//Self-reconciling version of the zitadel-operator
	Version string `yaml:"version,omitempty"`
}

func ParseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{Common: desiredTree.Common}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredKind.Common.Version = "v0"

	return desiredKind, nil
}
