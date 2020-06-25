package certificate

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, namespace string, clients []string) (cockroachdb.QueryFunc, cockroachdb.DestroyFunc, error) {
	secrets, err := k8sClient.ListSecrets(namespace)
	if err != nil {
		return nil, nil, err
	}

	return func() (cockroachdb.EnsureFunc, error) {
			caPrivKey, caCert, err := NewCA()
			if err != nil {
				return nil, err
			}

			applySecrets := make([]*corev1.Secret, 0)
			nodeSecret := "cockroachdb.node"
			found := false
			for _, secret := range secrets.Items {
				if secret.Name == nodeSecret {
					found = true
				}
			}

			if !found {
				nodePrivKey, nodeCert, err := NewNode(caPrivKey, caCert)
				if err != nil {
					return nil, err
				}

				pemNodePrivKey, err := PEMEncodeKey(nodePrivKey)
				if err != nil {
					return nil, err
				}

				caCertKey := "ca.crt"
				nodeCertKey := "node.crt"
				nodePrivKeyKey := "node.key"

				applySecrets = append(applySecrets, &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      nodeSecret,
						Namespace: namespace,
					},
					Type: corev1.SecretTypeOpaque,
					StringData: map[string]string{
						caCertKey:      string(caCert),
						nodePrivKeyKey: string(pemNodePrivKey),
						nodeCertKey:    string(nodeCert),
					},
				})
			}

			for _, client := range clients {
				found := false
				for _, secret := range secrets.Items {
					if secret.Name == client {
						found = true
					}
				}

				if !found {
					clientPrivKey, clientCert, err := NewClient(caPrivKey, caCert, client)
					if err != nil {
						return nil, err
					}

					pemClientPrivKey, err := PEMEncodeKey(clientPrivKey)
					if err != nil {
						return nil, err
					}

					caCertKey := "ca.crt"
					clientCertKey := "client" + client + ".crt"
					clientPrivKeyKey := "client." + client + ".key"

					applySecrets = append(applySecrets, &corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      client,
							Namespace: namespace,
						},
						Type: corev1.SecretTypeOpaque,
						StringData: map[string]string{
							caCertKey:        string(caCert),
							clientPrivKeyKey: string(pemClientPrivKey),
							clientCertKey:    string(clientCert),
						},
					})
				}
			}

			return func() error {
				for _, secret := range applySecrets {
					if err := k8sClient.ApplySecret(secret); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
