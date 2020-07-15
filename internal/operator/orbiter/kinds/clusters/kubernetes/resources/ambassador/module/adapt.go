package module

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Config struct {
	EnableGrpcWeb bool
}

func AdaptFunc(name, namespace string, labels map[string]string, config *Config) (resources.QueryFunc, resources.DestroyFunc, error) {
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

	return func(k8sClient *kubernetes.Client) (resources.EnsureFunc, error) {
			res, err := k8sClient.GetNamespacedCRDResource(group, version, kind, namespace, name)
			if err != nil && !macherrs.IsNotFound(err) {
				return nil, err
			}
			resourceVersion := ""
			if res != nil {
				meta := res.Object["metadata"].(map[string]interface{})
				resourceVersion = meta["resourceVersion"].(string)
			}

			if resourceVersion != "" {
				meta := crd.Object["metadata"].(map[string]interface{})
				meta["resourceVersion"] = resourceVersion
			}

			return func(k8sClient *kubernetes.Client) error {
				return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
