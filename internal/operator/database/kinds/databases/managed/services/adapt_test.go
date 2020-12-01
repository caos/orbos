package services

import (
	"github.com/caos/orbos/mntr"
	kubernetesmock "github.com/caos/orbos/pkg/kubernetes/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestService_Adapt1(t *testing.T) {
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	monitor := mntr.Monitor{}
	namespace := "testNs"
	publicService := "testPublic"
	service := "testSvc"
	labels := map[string]string{"test": "test"}
	publicLabels := map[string]string{"test": "test", "database.caos.ch/servicetype": "public"}
	internalLabels := map[string]string{"test": "test", "database.caos.ch/servicetype": "internal"}
	cockroachPort := int32(25267)
	cockroachHttpPort := int32(8080)
	queried := map[string]interface{}{}

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      publicService,
			Namespace: namespace,
			Labels:    publicLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: false,
		},
	})

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      publicService,
			Namespace: "default",
			Labels:    publicLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: false,
		},
	})

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service,
			Namespace: namespace,
			Labels:    internalLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
		},
	})

	query, _, err := AdaptFunc(monitor, namespace, publicService, service, labels, cockroachPort, cockroachHttpPort)
	assert.NoError(t, err)

	ensure, err := query(k8sClient, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)

	assert.NoError(t, ensure(k8sClient))
}

func TestService_Adapt2(t *testing.T) {
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	monitor := mntr.Monitor{}
	namespace := "testNs2"
	publicService := "testPublic2"
	service := "testSvc2"
	labels := map[string]string{"test2": "test2"}
	publicLabels := map[string]string{"test2": "test2", "database.caos.ch/servicetype": "public"}
	internalLabels := map[string]string{"test2": "test2", "database.caos.ch/servicetype": "internal"}
	cockroachPort := int32(23)
	cockroachHttpPort := int32(24)
	queried := map[string]interface{}{}

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      publicService,
			Namespace: namespace,
			Labels:    publicLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: false,
		},
	})

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      publicService,
			Namespace: "default",
			Labels:    publicLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: false,
		},
	})

	k8sClient.EXPECT().ApplyService(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service,
			Namespace: namespace,
			Labels:    internalLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 26257, TargetPort: intstr.FromInt(int(cockroachPort)), Name: "grpc"},
				{Port: 8080, TargetPort: intstr.FromInt(int(cockroachHttpPort)), Name: "http"},
			},
			Selector:                 labels,
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
		},
	})

	query, _, err := AdaptFunc(monitor, namespace, publicService, service, labels, cockroachPort, cockroachHttpPort)
	assert.NoError(t, err)

	ensure, err := query(k8sClient, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)

	assert.NoError(t, ensure(k8sClient))
}
