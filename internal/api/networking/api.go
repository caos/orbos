package networking

import (
	"errors"

	v1 "github.com/caos/orbos/internal/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"sigs.k8s.io/yaml"
)

const (
	namespace  = "caos-system"
	kind       = "Networking"
	apiVersion = "caos.ch/v1"
	name       = "networking"
)

func ReadCRD(k8sClient kubernetes.ClientInt) (*tree.Tree, error) {

	unstruct, err := k8sClient.GetNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, kind, namespace, name)
	if err != nil {
		if macherrs.IsNotFound(err) || meta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, err
	}

	spec, found := unstruct.Object["spec"]
	if !found {
		return nil, errors.New("no spec in crd")
	}
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return nil, errors.New("no spec in crd")
	}

	data, err := yaml.Marshal(specMap)
	if err != nil {
		return nil, err
	}

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

	unstruct.SetName(name)
	unstruct.SetNamespace(namespace)
	unstruct.SetKind(kind)
	unstruct.SetAPIVersion(apiVersion)

	return k8sClient.ApplyNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, kind, namespace, name, unstruct)
}