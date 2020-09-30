package module

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Config struct {
	EnableGrpcWeb bool
}

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Module"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string, config *Config) (resources.QueryFunc, error) {
	spec := map[string]interface{}{}
	if config != nil {
		specConfig := map[string]interface{}{}
		if config.EnableGrpcWeb {
			specConfig["enable_grpc_web"] = config.EnableGrpcWeb
		}
		spec["config"] = specConfig
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

	return func(k8sClient *kubernetes2.Client) (resources.EnsureFunc, error) {
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

		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
