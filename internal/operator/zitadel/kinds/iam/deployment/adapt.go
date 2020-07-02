package deployment

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/deployment"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type secret struct {
	Path      string
	Name      string
	Namespace string
}

func AdaptFunc(
	namespace string,
	labels map[string]string,
	replicaCount int,
	version string,
) (
	func(currentDB interface{}) (resources.EnsureFunc, error),
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
							Image:           "docker.pkg.github.com/caos/zitadel/zitadel:" + version,
							ImagePullPolicy: "IfNotPresent",
							Ports: []v1.ContainerPort{
								{Name: "management-rest", ContainerPort: 50011},
								{Name: "management-grpc", ContainerPort: 50010},
								{Name: "auth-rest", ContainerPort: 50021},
								{Name: "issuer-rest", ContainerPort: 50022},
								{Name: "auth-grpc", ContainerPort: 50020},
								{Name: "admin-rest", ContainerPort: 50041},
								{Name: "admin-grpc", ContainerPort: 50040},
								{Name: "console-http", ContainerPort: 50050},
								{Name: "accounts-http", ContainerPort: 50031},
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
						Name: "public-github-packages",
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

	_, destroy, err := deployment.AdaptFunc(deploymentDef)
	if err != nil {
		return nil, nil, err
	}

	return func(currentDB interface{}) (resources.EnsureFunc, error) {
		current := currentDB.(core.DatabaseCurrent)
		secrets := make([]*secret, 0)
		users := current.GetUsers()
		for _, user := range users {
			secrets = append(secrets, &secret{
				Path: "/certs/" + user,
				Name: "cockroachdb.client." + user,
			})
		}

		for _, secret := range secrets {
			internalName := strings.ReplaceAll(secret.Name, ".", "-")
			vol := v1.Volume{
				Name: internalName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secret.Name,
					},
				},
			}
			deploymentDef.Spec.Template.Spec.Volumes = append(deploymentDef.Spec.Template.Spec.Volumes, vol)

			volMount := v1.VolumeMount{Name: internalName, MountPath: secret.Path}
			deploymentDef.Spec.Template.Spec.Containers[0].VolumeMounts = append(deploymentDef.Spec.Template.Spec.Containers[0].VolumeMounts, volMount)
		}

		query, _, err := deployment.AdaptFunc(deploymentDef)
		if err != nil {
			return nil, err
		}
		return query()
	}, destroy, nil
}
