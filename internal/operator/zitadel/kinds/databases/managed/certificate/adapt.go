package certificate

import (
	"crypto/rsa"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	"strings"
)

func AdaptFunc(namespace string, clients []string, labels map[string]string, clusterDns string) (zitadel.QueryFunc, zitadel.DestroyFunc, error) {
	nodeLabels := map[string]string{}
	clientLabels := map[string]string{}
	for k, v := range labels {
		nodeLabels[k] = v
		clientLabels[k] = v
	}
	nodeLabels["zitadel.caos.ch/secret-type"] = "node"
	clientLabels["zitadel.caos.ch/secret-type"] = "client"

	desiredNode := make([]*secretInternal, 0)
	desiredClient := make([]*secretInternal, 0)

	desiredNode = append(desiredNode, &secretInternal{
		name:      "nodeSecret",
		namespace: "namespace",
		labels:    nodeLabels,
	})

	for _, client := range clients {
		clientSecret := "cockroachdb.client." + client
		desiredClient = append(desiredClient, &secretInternal{
			client:    client,
			name:      strings.ReplaceAll(clientSecret, "_", "-"),
			namespace: namespace,
			labels:    clientLabels,
		})
	}

	return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
			ensurers := make([]resources.EnsureFunc, 0)

			allNodeSecrets, err := k8sClient.ListSecrets(namespace, nodeLabels)
			if err != nil {
				return nil, err
			}

			caPrivKey := new(rsa.PrivateKey)
			caCert := make([]byte, 0)
			nodeSecret := "cockroachdb.node"
			caCertKey := "ca.crt"
			caPrivKeyKey := "ca.key"
			nodeCertKey := "node.crt"
			nodePrivKeyKey := "node.key"

			if len(allNodeSecrets.Items) == 0 {
				caPrivKeyInternal, caCertInternal, err := NewCA()
				if err != nil {
					return nil, err
				}
				caPrivKey = caPrivKeyInternal
				caCert = caCertInternal

				nodePrivKey, nodeCert, err := NewNode(caPrivKey, caCert, namespace, clusterDns)
				if err != nil {
					return nil, err
				}

				pemNodePrivKey, err := PEMEncodeKey(nodePrivKey)
				if err != nil {
					return nil, err
				}
				pemCaPrivKey, err := PEMEncodeKey(caPrivKey)
				if err != nil {
					return nil, err
				}

				pemCaCert, err := PEMEncodeCertificate(caCert)
				if err != nil {
					return nil, err
				}

				pemNodeCert, err := PEMEncodeCertificate(nodeCert)
				if err != nil {
					return nil, err
				}

				nodeSecretData := map[string]string{
					caPrivKeyKey:   string(pemCaPrivKey),
					caCertKey:      string(pemCaCert),
					nodePrivKeyKey: string(pemNodePrivKey),
					nodeCertKey:    string(pemNodeCert),
				}
				queryNodeSecret, _, err := secret.AdaptFunc(nodeSecret, namespace, nodeLabels, nodeSecretData)
				if err != nil {
					return nil, err
				}
				ensure, err := queryNodeSecret(k8sClient)
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, ensure)
			} else {
				cert, err := PEMDecodeKey(allNodeSecrets.Items[0].Data[caPrivKeyKey])
				if err != nil {
					return nil, err
				}
				caPrivKey = cert

				caCert = allNodeSecrets.Items[0].Data[caCertKey]
			}

			allClientSecrets, err := k8sClient.ListSecrets(namespace, clientLabels)
			if err != nil {
				return nil, err
			}
			createSecrets, deleteSecrets := queryCertificate(desiredClient, allClientSecrets.Items)

			for _, createSecret := range createSecrets {
				clientPrivKey, clientCert, err := NewClient(caPrivKey, caCert, createSecret.client)
				if err != nil {
					return nil, err
				}

				pemClientPrivKey, err := PEMEncodeKey(clientPrivKey)
				if err != nil {
					return nil, err
				}

				pemClientCert, err := PEMEncodeCertificate(clientCert)
				if err != nil {
					return nil, err
				}

				pemCaCert, err := PEMEncodeCertificate(caCert)
				if err != nil {
					return nil, err
				}

				caCertKey := "ca.crt"
				clientCertKey := "client." + createSecret.client + ".crt"
				clientPrivKeyKey := "client." + createSecret.client + ".key"

				clientSecretData := map[string]string{
					caCertKey:        string(pemCaCert),
					clientPrivKeyKey: string(pemClientPrivKey),
					clientCertKey:    string(pemClientCert),
				}

				queryClientSecret, _, err := secret.AdaptFunc(createSecret.name, namespace, clientLabels, clientSecretData)
				if err != nil {
					return nil, err
				}

				ensure, err := queryClientSecret(k8sClient)
				if err != nil {
					return nil, err
				}

				ensurers = append(ensurers, ensure)
			}
			for _, deleteSecret := range deleteSecrets {
				_, destroy, err := secret.AdaptFunc(deleteSecret.name, namespace, clientLabels, map[string]string{})
				if err != nil {
					return nil, err
				}

				ensurers = append(ensurers, func(client *kubernetes.Client) error {
					return destroy(client)
				})
			}

			return func(k8sClient *kubernetes.Client) error {
				for _, ensurer := range ensurers {
					if err := ensurer(k8sClient); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			allClientSecrets, err := k8sClient.ListSecrets(namespace, clientLabels)
			if err != nil {
				return err
			}
			_, deleteSecrets := queryCertificate([]*secretInternal{}, allClientSecrets.Items)
			for _, deleteSecret := range deleteSecrets {
				_, destroyer, err := secret.AdaptFunc(deleteSecret.name, namespace, clientLabels, map[string]string{})
				if err != nil {
					return err
				}
				if err := destroyer(k8sClient); err != nil {
					return err
				}
			}
			return nil
		}, nil
}
