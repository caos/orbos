package module

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
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

	annotations := map[string]string{}
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": group + "/" + version,
			"metadata": map[string]interface{}{
				"name":        name,
				"namespace":   namespace,
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": spec,
		}}

	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {
		existing, err := k8sClient.GetNamespacedCRDResource(group, version, kind, namespace, name)
		if err != nil && !macherrs.IsNotFound(err) {
			return nil, err
		}

		if !macherrs.IsNotFound(err) {
			exisistingLabels := make(map[string]string)
			exisistingAnnotations := make(map[string]string)
			metadataT, ok := existing.Object["metadata"]
			if ok && metadataT != nil {
				existingMetadata := metadataT.(map[string]interface{})

				labelsT, ok := existingMetadata["labels"]
				if ok && labelsT != nil {
					exisistingLabels = labelsT.(map[string]string)
				}

				annotationsT, ok := existingMetadata["annotations"]
				if ok && annotationsT != nil {
					exisistingAnnotations = annotationsT.(map[string]string)
				}
			}

			if !reflect.DeepEqual(labels, exisistingLabels) ||
				!reflect.DeepEqual(annotations, exisistingAnnotations) ||
				!reflect.DeepEqual(crd.Object["spec"], existing.Object["spec"]) {
				return func(k8sClient kubernetes.ClientInt) error {
					return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
				}, nil
			}
			return func(k8sClient kubernetes.ClientInt) error {
				return nil
			}, nil
		}
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
