package client

import (
	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/certificates"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/pem"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/secret"
	"strings"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
) (
	func(client string) core2.QueryFunc,
	func(client string) core2.DestroyFunc,
	error,
) {
	clientLabels := map[string]string{}
	for k, v := range labels {
		clientLabels[k] = v
	}
	clientLabels["zitadel.caos.ch/secret-type"] = "client"

	clientSecretPrefix := "cockroachdb.client."

	return func(client string) core2.QueryFunc {
			clientSecret := clientSecretPrefix + client
			desiredClient := &certificates.SecretInternal{
				Client:    client,
				Name:      strings.ReplaceAll(clientSecret, "_", "-"),
				Namespace: namespace,
				Labels:    clientLabels,
			}

			return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (core2.EnsureFunc, error) {
				queriers := make([]core2.QueryFunc, 0)

				currentDB, err := core.ParseQueriedForDatabase(queried)
				if err != nil {
					return nil, err
				}

				clientPrivKey, clientCert, err := certificates.NewClient(currentDB.GetCertificateKey(), currentDB.GetCertificate(), desiredClient.Client)
				if err != nil {
					return nil, err
				}

				pemClientPrivKey, err := pem.EncodeKey(clientPrivKey)
				if err != nil {
					return nil, err
				}

				pemClientCert, err := pem.EncodeCertificate(clientCert)
				if err != nil {
					return nil, err
				}

				pemCaCert, err := pem.EncodeCertificate(currentDB.GetCertificate())
				if err != nil {
					return nil, err
				}

				caCertKey := "ca.crt"
				clientCertKey := "client." + desiredClient.Client + ".crt"
				clientPrivKeyKey := "client." + desiredClient.Client + ".key"

				clientSecretData := map[string]string{
					caCertKey:        string(pemCaCert),
					clientPrivKeyKey: string(pemClientPrivKey),
					clientCertKey:    string(pemClientCert),
				}

				queryClientSecret, err := secret.AdaptFuncToEnsure(namespace, desiredClient.Name, clientLabels, clientSecretData)
				if err != nil {
					return nil, err
				}
				queriers = append(queriers, core2.ResourceQueryToZitadelQuery(queryClientSecret))

				return core2.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
			}
		}, func(client string) core2.DestroyFunc {
			clientSecret := clientSecretPrefix + client

			destroy, err := secret.AdaptFuncToDestroy(namespace, clientSecret)
			if err != nil {
				return nil
			}
			return core2.ResourceDestroyToZitadelDestroy(destroy)
		},
		nil
}
