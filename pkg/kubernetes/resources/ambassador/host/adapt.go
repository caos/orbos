package host

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
)

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Host"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string, hostname string, authority string, privateKeySecret string, selector map[string]string, tlsSecret string) (resources.QueryFunc, error) {
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
	annotations := map[string]string{
		"aes_res_changed": "true",
	}

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
			"spec": map[string]interface{}{
				"hostname":     hostname,
				"acmeProvider": acme,
				"ambassadorId": []string{
					"default",
				},
				"selector": map[string]interface{}{
					"matchLabels": selectorInternal,
				},
				"tlsSecret": map[string]interface{}{
					"name": tlsSecret,
				},
			},
		}}

	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {
		existing, err := k8sClient.GetNamespacedCRDResource(group, version, kind, namespace, name)
		if err != nil && !macherrs.IsNotFound(err) {
			return nil, err
		}

		if !macherrs.IsNotFound(err) {
			exisistingLabels := make(map[string]interface{})
			exisistingAnnotations := make(map[string]interface{})
			metadataT, ok := existing.Object["metadata"]
			if ok && metadataT != nil {
				existingMetadata := metadataT.(map[string]interface{})

				labelsT, ok := existingMetadata["labels"]
				if ok && labelsT != nil {
					exisistingLabels = labelsT.(map[string]interface{})
				}

				annotationsT, ok := existingMetadata["annotations"]
				if ok && annotationsT != nil {
					exisistingAnnotations = annotationsT.(map[string]interface{})
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
