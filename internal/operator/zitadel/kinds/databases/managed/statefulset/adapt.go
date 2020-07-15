package statefulset

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func AdaptFunc(
	namespace string,
	name string,
	image string,
	labels map[string]string,
	serviceAccountName string,
	replicaCount *int32,
	storageCapacity string,
	dbPort int32,
	httpPort int32,
	storageClass string,
	nodeSelector map[string]string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {
	defaultMode := int32(256)
	quantity, err := resource.ParseQuantity(storageCapacity)
	if err != nil {
		return nil, nil, err
	}

	joinList := make([]string, 0)
	for i := int32(0); i < *replicaCount; i++ {
		joinList = append(joinList, fmt.Sprintf("%s-%d.%s.%s:%d", name, i, name, namespace, dbPort))
	}
	joinListStr := strings.Join(joinList, ",")

	locality := "zone=" + namespace
	certPath := "/cockroach/cockroach-certs"
	datadirPath := "/cockroach/cockroach-data"
	joinExec := "exec /cockroach/cockroach start --logtostderr --certs-dir " + certPath + " --advertise-host $(hostname -f) --http-addr 0.0.0.0 --join " + joinListStr + " --locality " + locality + " --cache 25% --max-sql-memory 25%"
	datadirInternal := "datadir"
	certsInternal := "certs"

	statefulsetDef := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: name,
			Replicas:    replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					NodeSelector:       nodeSelector,
					ServiceAccountName: serviceAccountName,
					Affinity: &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{{
											Key:      "app",
											Operator: "In",
											Values: []string{
												"cockroachdb",
											}},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							}},
						},
					},
					Containers: []corev1.Container{{
						Name:            name,
						Image:           image,
						ImagePullPolicy: "IfNotPresent",
						Ports: []corev1.ContainerPort{
							{ContainerPort: dbPort, Name: "grpc"},
							{ContainerPort: httpPort, Name: "http"},
						},
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/health",
									Port:   intstr.Parse("http"),
									Scheme: "HTTPS",
								},
							},
							InitialDelaySeconds: 30,
							PeriodSeconds:       5,
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/health?ready=1",
									Port:   intstr.Parse("http"),
									Scheme: "HTTPS",
								},
							},
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							FailureThreshold:    2,
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      datadirInternal,
							MountPath: datadirPath,
						}, {
							Name:      certsInternal,
							MountPath: certPath,
						}},
						Env: []corev1.EnvVar{{
							Name:  "COCKROACH_CHANNEL",
							Value: "kubernetes-multiregion",
						}},
						Command: []string{
							"/bin/bash",
							"-ecx",
							joinExec,
						},
					}},
					Volumes: []corev1.Volume{{
						Name: datadirInternal,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: datadirInternal,
							},
						},
					}, {
						Name: certsInternal,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  "cockroachdb.node",
								DefaultMode: &defaultMode,
							},
						},
					}},
				},
			},
			PodManagementPolicy: appsv1.PodManagementPolicyType("Parallel"),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: "RollingUpdate",
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name: datadirInternal,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.PersistentVolumeAccessMode("ReadWriteOnce"),
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": quantity,
						},
					},
					StorageClassName: &storageClass,
				},
			}},
		},
	}
	return statefulset.AdaptFunc(statefulsetDef)
}
