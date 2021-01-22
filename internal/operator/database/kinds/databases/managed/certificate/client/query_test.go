package client

import (
	kubernetesmock "github.com/caos/orbos/pkg/kubernetes/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestClient_Query0(t *testing.T) {
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	clientLabels := map[string]string{"test": "test", "database.caos.ch/secret-type": "client"}
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))

	secretList := &corev1.SecretList{
		Items: []corev1.Secret{},
	}

	k8sClient.EXPECT().ListSecrets(namespace, clientLabels).Times(1).Return(secretList, nil)

	users, err := QueryCertificates(namespace, labels, k8sClient)
	assert.NoError(t, err)
	assert.Equal(t, users, []string{})
}

func TestClient_Query(t *testing.T) {
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	clientLabels := map[string]string{"test": "test", "database.caos.ch/secret-type": "client"}
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))

	secretList := &corev1.SecretList{
		Items: []corev1.Secret{{
			ObjectMeta: metav1.ObjectMeta{
				Name: clientSecretPrefix + "test",
			},
			Data: map[string][]byte{},
			Type: "Opaque",
		}},
	}

	k8sClient.EXPECT().ListSecrets(namespace, clientLabels).Times(1).Return(secretList, nil)

	users, err := QueryCertificates(namespace, labels, k8sClient)
	assert.NoError(t, err)
	assert.Contains(t, users, "test")
}

func TestClient_Query2(t *testing.T) {
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	clientLabels := map[string]string{"test": "test", "database.caos.ch/secret-type": "client"}
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))

	secretList := &corev1.SecretList{
		Items: []corev1.Secret{{
			ObjectMeta: metav1.ObjectMeta{
				Name: clientSecretPrefix + "test1",
			},
			Data: map[string][]byte{},
			Type: "Opaque",
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name: clientSecretPrefix + "test2",
			},
			Data: map[string][]byte{},
			Type: "Opaque",
		}},
	}

	k8sClient.EXPECT().ListSecrets(namespace, clientLabels).Times(1).Return(secretList, nil)

	users, err := QueryCertificates(namespace, labels, k8sClient)
	assert.NoError(t, err)
	assert.Equal(t, users, []string{"test1", "test2"})
}
