package certificate

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
)

func AdaptFunc(k8sClient *kubernetes.Client, namespace string, clients []string, labels map[string]string) (resources.QueryFunc, resources.DestroyFunc, error) {
	caPrivKey, caCert, err := NewCA()
	if err != nil {
		return nil, nil, err
	}

	querySecrets := make([]resources.QueryFunc, 0)
	destroySecrets := make([]resources.DestroyFunc, 0)
	nodeSecret := "cockroachdb.node"

	nodePrivKey, nodeCert, err := NewNode(caPrivKey, caCert, namespace)
	if err != nil {
		return nil, nil, err
	}

	pemNodePrivKey, err := PEMEncodeKey(nodePrivKey)
	if err != nil {
		return nil, nil, err
	}

	pemCaCert, err := PEMEncodeCertificate(caCert)
	if err != nil {
		return nil, nil, err
	}

	pemNodeCert, err := PEMEncodeCertificate(nodeCert)
	if err != nil {
		return nil, nil, err
	}

	caCertKey := "ca.crt"
	nodeCertKey := "node.crt"
	nodePrivKeyKey := "node.key"

	nodeSecretData := map[string]string{
		caCertKey:      string(pemCaCert),
		nodePrivKeyKey: string(pemNodePrivKey),
		nodeCertKey:    string(pemNodeCert),
	}
	queryNodeSecret, destroyNodeSecret, err := secret.AdaptFunc(k8sClient, nodeSecret, namespace, labels, nodeSecretData)
	if err != nil {
		return nil, nil, err
	}
	querySecrets = append(querySecrets, queryNodeSecret)
	destroySecrets = append(destroySecrets, destroyNodeSecret)

	for _, client := range clients {
		clientPrivKey, clientCert, err := NewClient(caPrivKey, caCert, client)
		if err != nil {
			return nil, nil, err
		}

		pemClientPrivKey, err := PEMEncodeKey(clientPrivKey)
		if err != nil {
			return nil, nil, err
		}

		pemClientCert, err := PEMEncodeCertificate(clientCert)
		if err != nil {
			return nil, nil, err
		}

		caCertKey := "ca.crt"
		clientCertKey := "client." + client + ".crt"
		clientPrivKeyKey := "client." + client + ".key"

		clientSecret := "cockroachdb.client." + client
		clientSecretData := map[string]string{
			caCertKey:        string(pemCaCert),
			clientPrivKeyKey: string(pemClientPrivKey),
			clientCertKey:    string(pemClientCert),
		}
		queryClientSecret, destroyClientSecret, err := secret.AdaptFunc(k8sClient, clientSecret, namespace, labels, clientSecretData)
		if err != nil {
			return nil, nil, err
		}
		querySecrets = append(querySecrets, queryClientSecret)
		destroySecrets = append(destroySecrets, destroyClientSecret)
	}

	return func() (resources.EnsureFunc, error) {
			ensurers := make([]resources.EnsureFunc, 0)
			for _, querySecret := range querySecrets {
				ensure, err := querySecret()
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, ensure)
			}

			return func() error {
				for _, ensurer := range ensurers {
					if err := ensurer(); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func() error {
			for _, destroyer := range destroySecrets {
				if err := destroyer(); err != nil {
					return err
				}
			}
			return nil
		}, nil
}
