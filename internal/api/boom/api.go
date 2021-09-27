package boom

import (
	v1 "github.com/caos/orbos/v5/internal/api/boom/v1"
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/tree"
	"gopkg.in/yaml.v3"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	Namespace = "caos-system"
	Name      = "boom"
)

func ReadCRD(k8sClient kubernetes.ClientInt) (*tree.Tree, error) {
	unstruct, err := k8sClient.GetNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, "Boom", Namespace, Name)
	if err != nil && !macherrs.IsNotFound(err) && !meta.IsNoMatchError(err) {
		return nil, err
	}
	err = nil

	var data []byte
	if unstruct == nil {
		unstruct = &unstructured.Unstructured{Object: v1.GetEmpty(Namespace, Name)}
		err = nil
	}

	dataInt, err := yaml.Marshal(unstruct.Object)
	if err != nil {
		return nil, err
	}
	data = dataInt

	desired := &tree.Tree{}
	return desired, yaml.Unmarshal(data, desired)
}

func WriteCrd(k8sClient kubernetes.ClientInt, t *tree.Tree) error {

	data, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	unstruct := &unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}

	if err := yaml.Unmarshal(data, unstruct.Object); err != nil {
		return err
	}

	return k8sClient.ApplyNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, "Boom", Namespace, Name, unstruct)
}
