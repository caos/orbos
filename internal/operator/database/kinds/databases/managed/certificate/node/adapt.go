package node

import (
	"crypto/rsa"
	"reflect"

	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/certificates"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/pem"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/secret"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	clusterDns string,
) (
	core2.QueryFunc,
	core2.DestroyFunc,
	error,
) {
	nodeLabels := map[string]string{}
	for k, v := range labels {
		nodeLabels[k] = v
	}
	nodeLabels["zitadel.caos.ch/secret-type"] = "node"

	desiredNode := make([]*certificates.SecretInternal, 0)

	desiredNode = append(desiredNode, &certificates.SecretInternal{
		Name:      "nodeSecret",
		Namespace: "namespace",
		Labels:    nodeLabels,
	})

	caPrivKey := new(rsa.PrivateKey)
	caCert := make([]byte, 0)
	nodeSecret := "cockroachdb.node"
	caCertKey := "ca.crt"
	caPrivKeyKey := "ca.key"
	nodeCertKey := "node.crt"
	nodePrivKeyKey := "node.key"

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (core2.EnsureFunc, error) {
			queriers := make([]core2.QueryFunc, 0)

			currentDB, err := core.ParseQueriedForDatabase(queried)
			if err != nil {
				return nil, err
			}

			allNodeSecrets, err := k8sClient.ListSecrets(namespace, nodeLabels)
			if err != nil {
				return nil, err
			}

			if len(allNodeSecrets.Items) == 0 {
				emptyCert := true
				emptyKey := true
				if currentCaCert := currentDB.GetCertificate(); currentCaCert != nil && len(currentCaCert) != 0 {
					emptyCert = false
					caCert = currentCaCert
				}
				if currentCaCertKey := currentDB.GetCertificateKey(); currentCaCertKey != nil && !reflect.DeepEqual(currentCaCertKey, &rsa.PrivateKey{}) {
					emptyKey = false
					caPrivKey = currentCaCertKey
				}

				if emptyCert || emptyKey {
					caPrivKeyInternal, caCertInternal, err := certificates.NewCA()
					if err != nil {
						return nil, err
					}
					caPrivKey = caPrivKeyInternal
					caCert = caCertInternal

					nodePrivKey, nodeCert, err := certificates.NewNode(caPrivKey, caCert, namespace, clusterDns)
					if err != nil {
						return nil, err
					}

					pemNodePrivKey, err := pem.EncodeKey(nodePrivKey)
					if err != nil {
						return nil, err
					}
					pemCaPrivKey, err := pem.EncodeKey(caPrivKey)
					if err != nil {
						return nil, err
					}

					pemCaCert, err := pem.EncodeCertificate(caCert)
					if err != nil {
						return nil, err
					}

					pemNodeCert, err := pem.EncodeCertificate(nodeCert)
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
					queriers = append(queriers, core2.ResourceQueryToZitadelQuery(queryNodeSecret))
				}
			} else {
				key, err := pem.DecodeKey(allNodeSecrets.Items[0].Data[caPrivKeyKey])
				if err != nil {
					return nil, err
				}
				caPrivKey = key

				cert, err := pem.DecodeCertificate(allNodeSecrets.Items[0].Data[caCertKey])
				if err != nil {
					return nil, err
				}
				caCert = cert
			}

			currentDB.SetCertificate(caCert)
			currentDB.SetCertificateKey(caPrivKey)

			return core2.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		}, func(k8sClient *kubernetes.Client) error {
			allClientSecrets, err := k8sClient.ListSecrets(namespace, nodeLabels)
			if err != nil {
				return err
			}
			_, deleteSecrets := certificates.QueryCertificate([]*certificates.SecretInternal{}, allClientSecrets.Items)
			for _, deleteSecret := range deleteSecrets {
				destroyer, err := secret.AdaptFuncToDestroy(namespace, deleteSecret.Name)
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
