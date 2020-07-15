package module

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Config struct {
	EnableGrpcWeb bool
}

func AdaptFunc(name, namespace string, labels map[string]string, config *Config) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {

			group := "getambassador.io"
			version := "v2"
			kind := "Module"

			spec := map[string]interface{}{}
			if config != nil && config.EnableGrpcWeb {
				spec["enable_grpc_web"] = config.EnableGrpcWeb
			}

			crd := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       kind,
					"apiVersion": group + "/" + version,
					"metadata": map[string]interface{}{
						"name":      name,
						"namespace": namespace,
						"labels":    labels,
					},
					"spec": spec,
				}}

			return func(k8sClient *kubernetes.Client) error {
				return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
