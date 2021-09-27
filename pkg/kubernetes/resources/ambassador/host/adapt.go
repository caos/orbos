package host

import (
	"reflect"

	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Host"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string, hostname string, authority string, privateKeySecret string, selector map[string]string, tlsSecret string) (resources.QueryFunc, error) {

	labelInterfaceValues := make(map[string]interface{})
	for k, v := range labels {
		labelInterfaceValues[k] = v
	}

	acme := map[string]interface{}{
		"authority": authority,
	}
	if privateKeySecret != "" {
		acme["privateKeySecret"] = map[string]interface{}{
			"name": privateKeySecret,
		}
	}

	selectorInterfaceValues := make(map[string]interface{}, 0)
	for k, v := range selector {
		selectorInterfaceValues[k] = v
	}

	spec := map[string]interface{}{
		"hostname": hostname,
		"selector": map[string]interface{}{
			"matchLabels": selectorInterfaceValues,
		},
		"ambassador_id": []interface{}{"default"},
		"acmeProvider":  acme,
	}

	if tlsSecret != "" {
		spec["tlsSecret"] = map[string]interface{}{
			"name": tlsSecret,
		}
	}

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": group + "/" + version,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels":    labelInterfaceValues,
				"annotations": map[string]interface{}{
					"aes_res_changed": "true",
				},
			},
			"spec": spec,
		}}

	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {

		ensure := func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
		}

		existing, err := k8sClient.GetNamespacedCRDResource(group, version, kind, namespace, name)
		if err != nil && !macherrs.IsNotFound(err) {
			return nil, err
		}
		err = nil

		if existing == nil {
			return ensure, nil
		}

		if contains(existing.Object, crd.Object) {
			// Noop
			return func(clientInt kubernetes.ClientInt) error { return nil }, nil
		}

		return ensure, nil
	}, nil
}

// The order matters!!
// TODO: Is this reusable?
func contains(set, subset map[string]interface{}) bool {

	if len(set) < len(subset) {
		return false
	}

	for k, subsetValue := range subset {
		setValue, ok := set[k]
		if !ok {
			return false
		}
		setValueMap, setValueIsMap := setValue.(map[string]interface{})
		subsetValueMap, subsetValueIsMap := subsetValue.(map[string]interface{})
		if setValueIsMap != subsetValueIsMap {
			return false
		}
		if subsetValueIsMap {
			if contains(setValueMap, subsetValueMap) {
				continue
			}
			return false
		}
		if !reflect.DeepEqual(setValue, subsetValue) {
			return false
		}
	}

	return true
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
