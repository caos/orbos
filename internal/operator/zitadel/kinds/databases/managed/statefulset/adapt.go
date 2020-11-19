package statefulset

import (
	"fmt"
	"sort"
	"strings"

	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/statefulset"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Affinity struct {
	key   string
	value string
}

type Affinitys []metav1.LabelSelectorRequirement

func (a Affinitys) Len() int           { return len(a) }
func (a Affinitys) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Affinitys) Less(i, j int) bool { return a[i].Key < a[j].Key }

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	name string,
	image string,
	labels map[string]string,
	serviceAccountName string,
	replicaCount int,
	storageCapacity string,
	dbPort int32,
	httpPort int32,
	storageClass string,
	nodeSelector map[string]string,
	tolerations []corev1.Toleration,
	resourcesSFS *k8s.Resources,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	zitadel.EnsureFunc,
	zitadel.EnsureFunc,
	error,
) {
	internalMonitor := monitor.WithField("component", "statefulset")

	defaultMode := int32(256)
	quantity, err := resource.ParseQuantity(storageCapacity)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	replicaCountParsed := int32(replicaCount)
	joinList := make([]string, 0)
	for i := int32(0); i < replicaCountParsed; i++ {
		joinList = append(joinList, fmt.Sprintf("%s-%d.%s.%s:%d", name, i, name, namespace, dbPort))
	}
	joinListStr := strings.Join(joinList, ",")

	locality := "zone=" + namespace
	certPath := "/cockroach/cockroach-certs"
	clientCertPath := "/cockroach/cockroach-client-certs"
	datadirPath := "/cockroach/cockroach-data"
	joinExec := "exec /cockroach/cockroach start --logtostderr --certs-dir " + certPath + " --advertise-host $(hostname -f) --http-addr 0.0.0.0 --join " + joinListStr + " --locality " + locality + " --cache 25% --max-sql-memory 25%"
	datadirInternal := "datadir"
	certsInternal := "certs"
	clientCertsInternal := "client-certs"

	affinity := Affinitys{}
	for k, v := range labels {
		affinity = append(affinity, metav1.LabelSelectorRequirement{
			Key:      k,
			Operator: metav1.LabelSelectorOpIn,
			Values: []string{
				v,
			}})
	}
	sort.Sort(affinity)

	internalResources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("100m"),
			"memory": resource.MustParse("512Mi"),
		},
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("100m"),
			"memory": resource.MustParse("512Mi"),
		},
	}

	if resourcesSFS != nil {
		internalResources = corev1.ResourceRequirements{}
		if resourcesSFS.Requests != nil {
			internalResources.Requests = resourcesSFS.Requests
		}
		if resourcesSFS.Limits != nil {
			internalResources.Limits = resourcesSFS.Limits
		}
	}

	statefulsetDef := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: name,
			Replicas:    &replicaCountParsed,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					NodeSelector:       nodeSelector,
					Tolerations:        tolerations,
					ServiceAccountName: serviceAccountName,
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: affinity,
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
						}, {
							Name:      clientCertsInternal,
							MountPath: clientCertPath,
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
						Resources: internalResources,
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
					}, {
						Name: clientCertsInternal,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  "cockroachdb.client.root",
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

	query, err := statefulset.AdaptFuncToEnsure(statefulsetDef)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	destroy, err := statefulset.AdaptFuncToDestroy(namespace, name)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	wrapedQuery, wrapedDestroy, err := resources.WrapFuncs(internalMonitor, query, destroy)
	checkDBRunning := func(k8sClient *kubernetes.Client) error {
		internalMonitor.Info("waiting for statefulset to be running")
		if err := k8sClient.WaitUntilStatefulsetIsReady(namespace, name, true, false, 60); err != nil {
			internalMonitor.Error(errors.Wrap(err, "error while waiting for statefulset to be running"))
			return err
		}
		internalMonitor.Info("statefulset is running")
		return nil
	}

	checkDBNotReady := func(k8sClient *kubernetes.Client) error {
		internalMonitor.Info("checking for statefulset to not be ready")
		if err := k8sClient.WaitUntilStatefulsetIsReady(namespace, name, true, true, 1); err != nil {
			internalMonitor.Info("statefulset is not ready")
			return nil
		}
		internalMonitor.Info("statefulset is ready")
		return errors.New("statefulset is ready")
	}

	ensureInit := func(k8sClient *kubernetes.Client) error {
		if err := checkDBRunning(k8sClient); err != nil {
			return err
		}

		if err := checkDBNotReady(k8sClient); err != nil {
			return nil
		}

		command := "/cockroach/cockroach init --certs-dir=" + clientCertPath + " --host=" + name + "-0." + name

		if err := k8sClient.ExecInPod(namespace, name+"-0", name, command); err != nil {
			return err
		}
		return nil
	}

	checkDBReady := func(k8sClient *kubernetes.Client) error {
		internalMonitor.Info("waiting for statefulset to be ready")
		if err := k8sClient.WaitUntilStatefulsetIsReady(namespace, name, true, true, 60); err != nil {
			internalMonitor.Error(errors.Wrap(err, "error while waiting for statefulset to be ready"))
			return err
		}
		internalMonitor.Info("statefulset is ready")
		return nil
	}

	return wrapedQuery, wrapedDestroy, ensureInit, checkDBReady, err
}
