package mapping

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strconv"
)

type CORS struct {
	Origins        string
	Methods        string
	Headers        string
	Credentials    bool
	ExposedHeaders string
	MaxAge         string
}

const (
	group   = "getambassador.io"
	version = "v2"
	kind    = "Mapping"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string, grpc bool, host, prefix, rewrite, service, timeoutMS, connectTimeoutMS string, cors *CORS) (resources.QueryFunc, error) {

	spec := map[string]interface{}{
		"host":    host,
		"rewrite": rewrite,
		"service": service,
	}
	if prefix != "" {
		spec["prefix"] = prefix
	}

	if timeoutMS != "" {
		toMSint, err := strconv.Atoi(timeoutMS)
		if err != nil {
			return nil, err
		}
		if timeoutMS != "" {
			spec["timeout_ms"] = toMSint
		}
	}
	if connectTimeoutMS != "" {
		ctoMSint, err := strconv.Atoi(connectTimeoutMS)
		if err != nil {
			return nil, err
		}
		if connectTimeoutMS != "" {
			spec["connect_timeout_ms"] = ctoMSint
		}
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

	return func(k8sClient kubernetes2.ClientInt) (resources.EnsureFunc, error) {
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

		return func(k8sClient kubernetes2.ClientInt) error {
			return k8sClient.ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes2.ClientInt) error {
		return client.DeleteNamespacedCRDResource(group, version, kind, namespace, name)
	}, nil
}
