package managed

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configuration for managed cockroachDB
	Spec Spec
	//List of configrations to backup cockroachDB
	Backups map[string]*tree.Tree `yaml:"backups,omitempty"`
}

type Spec struct {
	//Verbose flag to set debug-level to debug
	Verbose bool
	//Number of replicas for the cockroachDB statefulset
	ReplicaCount int `yaml:"replicaCount,omitempty"`
	//Capacity for the PVC for cockroachDB
	StorageCapacity string `yaml:"storageCapacity,omitempty"`
	//Storageclass for the PVC for cockroachDB
	StorageClass string `yaml:"storageClass,omitempty"`
	//Nodeselector for statefulset and migration jobs
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	//Tolerations for statefulset and migration jobs
	Tolerations []corev1.Toleration `yaml:"tolerations,omitempty"`
	//DNS entry used for cockroachDB certificates, should be the same use for the cluster-DNS-entry
	ClusterDns string `yaml:"clusterDNS,omitempty"`
	//Resource limits and requests for cockroachDB statefulset
	Resources *k8s.Resources `yaml:"resources,omitempty"`
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
