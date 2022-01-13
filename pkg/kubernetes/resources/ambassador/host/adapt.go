package host

import (
	"reflect"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Host"
)

type Arguments struct {
	Monitor          mntr.Monitor
	Namespace        string
	Name             string
	Labels           map[string]string
	Hostname         string
	Authority        string
	PrivateKeySecret string
	Selector         map[string]string
	TlsSecret        string
}

//type AdaptFuncToEnsureFunc func(monitor mntr.Monitor, namespace, name string, labels map[string]string, hostname string, authority string, privateKeySecret string, selector map[string]string, tlsSecret string) (resources.QueryFunc, error)

func AdaptFuncToEnsure(params *Arguments) (resources.QueryFunc, error) {

	labelInterfaceValues := make(map[string]interface{})
	for k, v := range params.Labels {
		labelInterfaceValues[k] = v
	}

	acme := map[string]interface{}{
		"authority": params.Authority,
	}
	if params.PrivateKeySecret != "" {
		acme["privateKeySecret"] = map[string]interface{}{
			"name": params.PrivateKeySecret,
		}
	}

	selectorInterfaceValues := make(map[string]interface{}, 0)
	for k, v := range params.Selector {
		selectorInterfaceValues[k] = v
	}

	spec := map[string]interface{}{
		"hostname": params.Hostname,
		"selector": map[string]interface{}{
			"matchLabels": selectorInterfaceValues,
		},
		"ambassador_id": []interface{}{"default"},
		"acmeProvider":  acme,
	}

	if params.TlsSecret != "" {
		spec["tlsSecret"] = map[string]interface{}{
			"name": params.TlsSecret,
		}
	}

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": group + "/" + version,
			"metadata": map[string]interface{}{
				"name":      params.Name,
				"namespace": params.Namespace,
				"labels":    labelInterfaceValues,
				"annotations": map[string]interface{}{
					"aes_res_changed": "true",
				},
			},
			"spec": spec,
		}}

	return func(k8sClient kubernetes.ClientInt, _ map[string]interface{}) (resources.EnsureFunc, error) {
		ensure := func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, params.Namespace, params.Name, crd)
		}
		crdName := "hosts.getambassador.io"
		_, ok, err := k8sClient.CheckCRD(crdName)
		if err != nil {
			return nil, err
		}
		if !ok {
			params.Monitor.WithField("name", crdName).Info("crd definition not found, skipping")
			return func(k8sClient kubernetes.ClientInt) error { return nil }, nil
		}

		existing, err := k8sClient.GetNamespacedCRDResource(group, version, kind, params.Namespace, params.Name)
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
