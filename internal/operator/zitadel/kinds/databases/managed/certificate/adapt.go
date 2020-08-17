package certificate

import (
	"crypto/rsa"
	"errors"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	"reflect"
	"strings"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	clients []string,
	labels map[string]string,
	clusterDns string,
	caCertificate string,
	caKey string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	cMonitor := monitor.WithField("component", "certificates")

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

	caPrivKey := new(rsa.PrivateKey)
	caCert := make([]byte, 0)
	nodeSecret := "cockroachdb.node"
	caCertKey := "ca.crt"
	caPrivKeyKey := "ca.key"
	nodeCertKey := "node.crt"
	nodePrivKeyKey := "node.key"

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			queriers := make([]zitadel.QueryFunc, 0)

			if caCertificate != "" || caKey != "" {
				if caCertificate == "" || caKey == "" {
					return nil, errors.New("CA certificate and key required")
				}
				cert, err := PEMDecodeKey([]byte(caKey))
				if err != nil {
					return nil, err
				}
				caPrivKey = cert

				caCert = []byte(caCertificate)
			} else {
				caPrivKeyInternal, caCertInternal, err := NewCA()
				if err != nil {
					return nil, err
				}
				caPrivKey = caPrivKeyInternal
				caCert = caCertInternal
			}

			allNodeSecrets, err := k8sClient.ListSecrets(namespace, nodeLabels)
			if err != nil {
				return nil, err
			}

			if len(allNodeSecrets.Items) == 0 {
				if caPrivKey == nil && string(caCert) != "" {
					caPrivKeyInternal, caCertInternal, err := NewCA()
					if err != nil {
						return nil, err
					}
					caPrivKey = caPrivKeyInternal
					caCert = caCertInternal
				}

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
				queryNodeSecret, err := secret.AdaptFuncToEnsure(namespace, nodeSecret, nodeLabels, nodeSecretData)
				if err != nil {
					return nil, err
				}
				queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(queryNodeSecret))
			} else {
				cert, err := PEMDecodeKey(allNodeSecrets.Items[0].Data[caPrivKeyKey])
				if err != nil {
					return nil, err
				}
				if !reflect.DeepEqual(*cert, *caPrivKey) || !reflect.DeepEqual(caCert, allNodeSecrets.Items[0].Data[caCertKey]) {
					//TODO
					return nil, errors.New("CA changed, not implemented yet")
				}
				/*
					caPrivKey = cert
					caCert = allNodeSecrets.Items[0].Data[caCertKey]*/
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

				queryClientSecret, err := secret.AdaptFuncToEnsure(namespace, createSecret.name, clientLabels, clientSecretData)
				if err != nil {
					return nil, err
				}
				queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(queryClientSecret))
			}
			for _, deleteSecret := range deleteSecrets {
				destroy, err := secret.AdaptFuncToDestroy(namespace, deleteSecret.name)
				if err != nil {
					return nil, err
				}

				queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(func(client *kubernetes.Client) (ensureFunc resources.EnsureFunc, err error) {
					return func(client *kubernetes.Client) error {
						return destroy(client)
					}, nil
				}))
			}

			return zitadel.QueriersToEnsureFunc(cMonitor, false, queriers, k8sClient, queried)
		}, func(k8sClient *kubernetes.Client) error {
			allClientSecrets, err := k8sClient.ListSecrets(namespace, clientLabels)
			if err != nil {
				return err
			}
			_, deleteSecrets := queryCertificate([]*secretInternal{}, allClientSecrets.Items)
			for _, deleteSecret := range deleteSecrets {
				destroyer, err := secret.AdaptFuncToDestroy(namespace, deleteSecret.name)
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
