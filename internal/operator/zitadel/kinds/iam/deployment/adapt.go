package deployment

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/deployment"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	replicaCount int,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {
	replicas := int32(replicaCount)

	internalSecrets := "zitadel-secret"
	internalConfig := "console-config"

	deploymentDef := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zitadel",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "zitadel",
							Image:           "docker.pkg.github.com/caos/zitadel/zitadel:latest",
							ImagePullPolicy: "IfNotPresent",
							Ports: []v1.ContainerPort{
								{Name: "management-rest", HostPort: 50011},
								{Name: "management-grpc", HostPort: 50010},
								{Name: "auth-rest", HostPort: 50021},
								{Name: "issuer-rest", HostPort: 50022},
								{Name: "auth-grpc", HostPort: 50020},
								{Name: "admin-rest", HostPort: 50041},
								{Name: "admin-grpc", HostPort: 50040},
								{Name: "console-http", HostPort: 50050},
								{Name: "accounts-http", HostPort: 50031},
							},
							Env: []v1.EnvVar{
								{Name: "POD_IP",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									}},
								{Name: "CHAT_URL",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "zitadel-secrets-vars"},
											Key:                  "ZITADEL_GOOGLE_CHAT_URL",
										},
									}},
								{Name: "TWILIO_TOKEN",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "zitadel-secrets-vars"},
											Key:                  "ZITADEL_TWILIO_AUTH_TOKEN",
										},
									}},
								{Name: "TWILIO_SERVICE_SID",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "zitadel-secrets-vars"},
											Key:                  "ZITADEL_TWILIO_SID",
										},
									}},
								{Name: "SMTP_PASSWORD",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: "zitadel-secrets-vars"},
											Key:                  "ZITADEL_EMAILAPPKEY",
										},
									}},
							},
							EnvFrom: []v1.EnvFromSource{
								{ConfigMapRef: &v1.ConfigMapEnvSource{
									LocalObjectReference: v1.LocalObjectReference{Name: "zitadel-vars"},
								}}},
							VolumeMounts: []v1.VolumeMount{
								{Name: internalSecrets, MountPath: "/secret"},
								{Name: internalConfig, MountPath: "/console/environment.json", SubPath: "environment.json"},
							},
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{{
						Name: "githubsecret",
					}},
					Volumes: []v1.Volume{{
						Name: internalSecrets,
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "zitadel-secret",
							},
						},
					}, {
						Name: internalConfig,
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{Name: "console-config"},
							},
						},
					}},
				},
			},
		},
	}

	return deployment.AdaptFunc(deploymentDef)
}
