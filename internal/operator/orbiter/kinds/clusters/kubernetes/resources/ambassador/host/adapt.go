package host

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CORS struct {
	Origins        string
	Methods        string
	Headers        string
	Credentials    bool
	ExposedHeaders string
	MaxAge         string
}

func AdaptFunc(name, namespace string, labels map[string]string, hostname string, authority string, privateKeySecret string, selector map[string]string, tlsSecret string) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {

			kind := "Host"
			group := "getambassador.io"
			version := "v2"

			acme := map[string]interface{}{
				"authority": authority,
			}
			if privateKeySecret != "" {
				acme["privateKeySecret"] = map[string]interface{}{
					"name": privateKeySecret,
				}
			}

			selectorInternal := make(map[string]interface{}, 0)
			for k, v := range selector {
				selectorInternal[k] = v
			}
			crd := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       kind,
					"apiVersion": group + "/" + version,
					"metadata": map[string]interface{}{
						"name":      name,
						"namespace": namespace,
						"labels":    labels,
						"annotations": map[string]interface{}{
							"aes_res_changed": "true",
						},
					},
					"spec": map[string]interface{}{
						"hostname":     hostname,
						"acmeProvider": acme,
						"selector": map[string]interface{}{
							"matchLabels": selectorInternal,
						},
						"tlsSecret": map[string]interface{}{
							"name": tlsSecret,
						},
					},
				}}

			return func(k8sClient *kubernetes.Client) error {
				return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
