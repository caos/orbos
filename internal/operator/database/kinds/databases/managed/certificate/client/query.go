package client

import (
	"strings"

	"github.com/caos/orbos/pkg/kubernetes"
)

func QueryCertificates(
	namespace string,
	labels map[string]string,
	k8sClient *kubernetes.Client,
) (
	[]string,
	error,
) {

	clientLabels := map[string]string{}
	for k, v := range labels {
		clientLabels[k] = v
	}
	clientLabels["databases.caos.ch/secret-type"] = "client"

	list, err := k8sClient.ListSecrets(namespace, clientLabels)
	if err != nil {
		return nil, err
	}
	certs := []string{}
	for _, secret := range list.Items {
		certs = append(certs, strings.TrimPrefix(secret.Name, "cockroachdb.client."))
	}

	return certs, nil
}
