package deployment

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/deployment"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	replicaCount int,
	version string,
	imagePullSecret string,
	cmName string,
	certPath string,
	secretName string,
	secretPath string,
	consoleCMName string,
	secretVarsName string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {
	rootSecret := "client-root"
	secretMode := int32(0777)
	replicas := int32(replicaCount)
	runAsUser := int64(1000)
	runAsNonRoot := true
	certMountPath := "/dbsecrets"

	internalLabels := make(map[string]string, 0)
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "iam"

	userList := []string{"management", "auth", "authz", "adminapi", "notification"}
	volumnes := []v1.Volume{{
		Name: secretName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}, {
		Name: rootSecret,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  "cockroachdb.client.root",
				DefaultMode: &secretMode,
			},
		},
	}, {
		Name: consoleCMName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{Name: consoleCMName},
			},
		},
	}}
	volMounts := []v1.VolumeMount{
		{Name: secretName, MountPath: secretPath},
		{Name: consoleCMName, MountPath: "/console/environment.json", SubPath: "environment.json"},
		{Name: rootSecret, MountPath: certMountPath + "/ca.crt", SubPath: "ca.crt"},
	}
	for _, user := range userList {
		userReplaced := strings.ReplaceAll(user, "_", "-")
		internalName := "client-" + userReplaced
		volumnes = append(volumnes, v1.Volume{
			Name: internalName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName:  "cockroachdb.client." + userReplaced,
					DefaultMode: &secretMode,
				},
			},
		})
		volMounts = append(volMounts, v1.VolumeMount{
			Name: internalName,
			//ReadOnly:  true,
			MountPath: certMountPath + "/client." + user + ".crt",
			SubPath:   "client." + user + ".crt",
		})
		volMounts = append(volMounts, v1.VolumeMount{
			Name: internalName,
			//ReadOnly:  true,
			MountPath: certMountPath + "/client." + user + ".key",
			SubPath:   "client." + user + ".key",
		})
	}
	deploymentDef := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zitadel",
			Namespace: namespace,
			Labels:    internalLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: internalLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: internalLabels,
				},

				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:    &runAsUser,
						RunAsNonRoot: &runAsNonRoot,
					},
					Containers: []v1.Container{
						{
							Lifecycle: &v1.Lifecycle{
								PostStart: &v1.Handler{
									Exec: &v1.ExecAction{
										// TODO: until proper fix of https://github.com/kubernetes/kubernetes/issues/2630
										Command: []string{"sh", "-c",
											"mkdir -p " + certPath + "/ && cp " + certMountPath + "/* " + certPath + "/ && chmod 400 " + certPath + "/*"},
									},
								},
							},
							SecurityContext: &v1.SecurityContext{
								RunAsUser:    &runAsUser,
								RunAsNonRoot: &runAsNonRoot,
							},
							//Command:         []string{"/bin/sh", "-c", "while true; do sleep 30; done;"},
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
											LocalObjectReference: v1.LocalObjectReference{Name: secretVarsName},
											Key:                  "ZITADEL_GOOGLE_CHAT_URL",
										},
									}},
								{Name: "TWILIO_TOKEN",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: secretVarsName},
											Key:                  "ZITADEL_TWILIO_AUTH_TOKEN",
										},
									}},
								{Name: "TWILIO_SERVICE_SID",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: secretVarsName},
											Key:                  "ZITADEL_TWILIO_SID",
										},
									}},
								{Name: "SMTP_PASSWORD",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: secretVarsName},
											Key:                  "ZITADEL_EMAILAPPKEY",
										},
									}},
							},
							EnvFrom: []v1.EnvFromSource{
								{ConfigMapRef: &v1.ConfigMapEnvSource{
									LocalObjectReference: v1.LocalObjectReference{Name: cmName},
								}}},
							VolumeMounts: volMounts,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{{
						Name: imagePullSecret,
					}},
					Volumes: volumnes,
				},
			},
		},
	}

	return deployment.AdaptFunc(deploymentDef)
}
