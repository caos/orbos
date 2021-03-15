package host

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
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
	acme := map[string]interface{}{
		"authority": params.Authority,
	}
	if params.PrivateKeySecret != "" {
		acme["privateKeySecret"] = map[string]interface{}{
			"name": params.PrivateKeySecret,
		}
	}

	selectorInternal := make(map[string]interface{}, 0)
	for k, v := range params.Selector {
		selectorInternal[k] = v
	}
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": group + "/" + version,
			"metadata": map[string]interface{}{
				"name":      params.Name,
				"namespace": params.Namespace,
				"labels":    params.Labels,
				"annotations": map[string]interface{}{
					"aes_res_changed": "true",
				},
			},
			"spec": map[string]interface{}{
				"hostname":     params.Hostname,
				"acmeProvider": acme,
				"ambassadorId": []string{
					"default",
				},
				"selector": map[string]interface{}{
					"matchLabels": selectorInternal,
				},
				"tlsSecret": map[string]interface{}{
					"name": params.TlsSecret,
				},
			},
		}}

	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {
		crdName := "hosts.getambassador.io"
		_, ok, err := k8sClient.CheckCRD(crdName)
		if err != nil {
			return nil, err
		}
		if !ok {
			params.Monitor.WithField("name", crdName).Info("crd definition not found, skipping")
			return func(k8sClient kubernetes.ClientInt) error { return nil }, nil
		}

		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, params.Namespace, params.Name, crd)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
