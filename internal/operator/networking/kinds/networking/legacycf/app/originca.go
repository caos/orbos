package app

import (
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/cloudflare/certificate"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/cloudflare/cloudflare-go"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func (a *App) EnsureOriginCACertificate(k8sClient *kubernetes2.Client, namespace string, labels map[string]string, domain string, originCASecretName string) error {

	certKey := "tls.crt"
	keyKey := "tls.key"

	secretList, err := k8sClient.ListSecrets(namespace, labels)
	if err != nil {
		return err
	}

	tlsSecret := new(corev1.Secret)
	for _, secret := range secretList.Items {
		if secret.Name == originCASecretName {
			tlsSecret = &secret
		}
	}

	caHosts := []string{
		"*." + domain,
		domain,
	}

	current, err := a.cloudflare.GetOriginCACertificates(domain)
	if err != nil {
		return err
	}

	foundCA := new(cloudflare.OriginCACertificate)
	for _, currentCA := range current {
		if reflect.DeepEqual(caHosts, currentCA.Hostnames) {
			foundCA = &currentCA
		}
	}

	ensured := false
	if foundCA != nil && tlsSecret != nil {
		data, ok := tlsSecret.Data[certKey]
		if ok && foundCA.Certificate == string(data) {
			ensured = true
		}
	}

	if !ensured {
		if foundCA != nil && foundCA.ID != "" {
			if err := a.cloudflare.RevokeOriginCACertificate(foundCA.ID); err != nil {
				return err
			}
		}
		priv, err := certificate.CreatePrivateKey()
		if err != nil {
			return err
		}

		origin, err := a.cloudflare.CreateOriginCACertificate(domain, caHosts, priv)
		if err != nil {
			return err
		}
		keyPem, err := certificate.PEMEncodeKey(priv)

		if err := k8sClient.ApplySecret(&corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      originCASecretName,
				Namespace: namespace,
				Labels:    labels,
			},
			StringData: map[string]string{
				certKey: origin.Certificate,
				keyKey:  string(keyPem),
			},
			Type: corev1.SecretTypeOpaque,
		}); err != nil {
			return err
		}
	}

	return nil
}
