package mapping

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
)

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Mapping"
)

type CORS struct {
	Origins        string
	Methods        string
	Headers        string
	Credentials    bool
	ExposedHeaders string
	MaxAge         string
}

func AdaptFuncToEnsure(
	namespace string,
	id labels.IDLabels,
	grpc bool,
	host,
	prefix,
	rewrite,
	service string,
	timeoutMS,
	connectTimeoutMS int,
	cors *CORS) (resources.QueryFunc, error) {

	spec := map[string]interface{}{
		"host":    host,
		"rewrite": rewrite,
		"service": service,
	}
	if prefix != "" {
		spec["prefix"] = prefix
	}

	if timeoutMS != 0 {
		spec["timeout_ms"] = timeoutMS
	}
	if connectTimeoutMS != 0 {
		spec["connect_timeout_ms"] = connectTimeoutMS
	}
	if grpc {
		spec["grpc"] = grpc
	}

	if cors != nil {
		corsMap := map[string]interface{}{
			"origins":         cors.Origins,
			"methods":         cors.Methods,
			"headers":         cors.Headers,
			"credentials":     cors.Credentials,
			"exposed_headers": cors.ExposedHeaders,
			"max_age":         cors.MaxAge,
		}
		spec["cors"] = corsMap
	}
	crdlabels := labels.MustK8sMap(id)
	annotations := map[string]string{}

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": group + "/" + version,
			"metadata": map[string]interface{}{
				"name":        id.Name(),
				"namespace":   namespace,
				"labels":      crdlabels,
				"annotations": annotations,
			},
			"spec": spec,
		}}

	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {
		existing, err := k8sClient.GetNamespacedCRDResource(group, version, kind, namespace, id.Name())
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

			if !reflect.DeepEqual(crdlabels, exisistingLabels) ||
				!reflect.DeepEqual(annotations, exisistingAnnotations) ||
				!reflect.DeepEqual(crd.Object["spec"], existing.Object["spec"]) {
				return func(k8sClient kubernetes.ClientInt) error {
					return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, id.Name(), crd)
				}, nil
			}
			return func(k8sClient kubernetes.ClientInt) error {
				return nil
			}, nil
		}
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, id.Name(), crd)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
