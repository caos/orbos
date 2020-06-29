package db

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrole"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/namespace"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/pdb"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/role"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/rolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/serviceaccount"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/statefulset"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/certificate"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func AdaptFunc() cockroachdb.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		cockroachdb.QueryFunc,
		cockroachdb.DestroyFunc,
		error,
	) {
		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		data, err := ioutil.ReadFile("/Users/benz/.kube/config")
		dummyKubeconfig := string(data)
		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		//if err := k8sClient.RefreshLocal(); err != nil {
		//	return nil, nil, err
		//}

		if !k8sClient.Available() {
			return nil, nil, errors.New("kubeconfig failed")
		}
		queriers := make([]resources.QueryFunc, 0)
		destroyers := make([]resources.DestroyFunc, 0)

		namespaceStr := "caos-cockroach"
		labels := map[string]string{"app.kubernetes.io/managed-by": "zitadel.caos.ch"}
		serviceAccountName := "cockroachdb"
		roleName := "cockroachdb"
		clusterRoleName := "cockroachdb"

		queryNS, destroyNS, err := namespace.AdaptFunc(k8sClient, namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		userList := []string{"root"}
		if desiredKind.Spec.Users != nil && len(desiredKind.Spec.Users) > 0 {
			userList = append(userList, desiredKind.Spec.Users...)
		}
		queryCert, destroyCert, err := certificate.AdaptFunc(k8sClient, namespaceStr, userList, labels)
		if err != nil {
			return nil, nil, err
		}

		querySA, destroySA, err := serviceaccount.AdaptFunc(k8sClient, namespaceStr, serviceAccountName, labels)
		if err != nil {
			return nil, nil, err
		}

		queryR, destroyR, err := role.AdaptFunc(k8sClient, roleName, namespaceStr, labels, []string{""}, []string{"secrets"}, []string{"create", "get"})
		if err != nil {
			return nil, nil, err
		}

		queryCR, destroyCR, err := clusterrole.AdaptFunc(k8sClient, clusterRoleName, labels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
		if err != nil {
			return nil, nil, err
		}

		subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryRB, destroyRB, err := rolebinding.AdaptFunc(k8sClient, roleName, namespaceStr, labels, subjects, roleName)
		if err != nil {
			return nil, nil, err
		}

		subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryCRB, destroyCRB, err := clusterrolebinding.AdaptFunc(k8sClient, roleName, labels, subjectsCRB, roleName)
		if err != nil {
			return nil, nil, err
		}

		ports := []service.Port{
			{
				Port:       26257,
				TargetPort: "26257",
				Name:       "grpc",
			},
			{
				Port:       8080,
				TargetPort: "8080",
				Name:       "http",
			},
		}
		querySP, destroySP, err := service.AdaptFunc(k8sClient, "cockroachdb-public", namespaceStr, labels, ports, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := service.AdaptFunc(k8sClient, "cockroachdb", namespaceStr, labels, ports, "", labels, true, "None", "")
		if err != nil {
			return nil, nil, err
		}

		replicas := int32(desiredKind.Spec.ReplicaCount)
		defaultMode := int32(256)
		quantity, err := resource.ParseQuantity(desiredKind.Spec.StorageCapacity)
		if err != nil {
			return nil, nil, err
		}
		joinList := make([]string, replicas)
		for i := int32(0); i < replicas; i++ {
			joinList = append(joinList, fmt.Sprintf("cockroachdb-%d.cockroachdb.%s", i, namespaceStr))
		}
		joinListStr := strings.Join(joinList, ",")
		locality := "zone=" + namespaceStr
		certPath := "/cockroach/cockroach-certs"
		joinExec := "exec /cockroach/cockroach start --logtostderr --certs-dir " + certPath + " --advertise-host $(hostname -f) --http-addr 0.0.0.0 --join " + joinListStr + " --locality " + locality + " --cache 25% --max-sql-memory 25%"

		statefulsetDef := &appsv1.StatefulSet{
			ObjectMeta: v1.ObjectMeta{
				Name:      "cockroachdb",
				Namespace: namespaceStr,
				Labels:    labels,
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: "cockroachdb",
				Replicas:    &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: serviceAccountName,
						Affinity: &corev1.Affinity{
							PodAffinity: &corev1.PodAffinity{
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &v1.LabelSelector{
											MatchExpressions: []v1.LabelSelectorRequirement{{
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
							Name:            "cockroachdb",
							Image:           "cockroachdb/cockroach:v20.1.2",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 26257, Name: "grpc"},
								{ContainerPort: 8080, Name: "http"},
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
								Name:      "datadir",
								MountPath: "/cockroach/cockroach-data",
							}, {
								Name:      "certs",
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
							Name: "datadir",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "datadir",
								},
							},
						}, {
							Name: "certs",
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
					ObjectMeta: v1.ObjectMeta{
						Name: "datadir",
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
					},
				}},
			},
		}

		querySFS, destroySFS, err := statefulset.AdaptFunc(k8sClient, statefulsetDef)
		if err != nil {
			return nil, nil, err
		}

		queryPDB, destroyPDB, err := pdb.AdaptFunc(k8sClient, namespaceStr, "cockroachdb-budget", labels, "1")
		if err != nil {
			return nil, nil, err
		}

		externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		queryES, destroyES, err := service.AdaptFunc(k8sClient, "cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		if err != nil {
			return nil, nil, err
		}

		jobDef := &batch.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "cockroachdb-cluster-init",
				Namespace: namespaceStr,
				Labels:    labels,
			},
			Spec: batch.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						ServiceAccountName: serviceAccountName,
						Containers: []corev1.Container{{
							Name:            "cockroachdb-cluster-init",
							Image:           "cockroachdb/cockroach:v20.1.2",
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "client-certs",
								MountPath: certPath,
							}},
							Command: []string{
								"/cockroach/cockroach",
								"init",
								"--certs-dir=" + certPath,
								"--host=cockroachdb-0.cockroachdb",
							},
						}},
						RestartPolicy: "OnFailure",
						Volumes: []corev1.Volume{{
							Name: "client-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "cockroachdb.client.root",
									DefaultMode: &defaultMode,
								},
							},
						}},
					},
				},
			},
		}
		queryJ, destroyJ, err := job.AdaptFunc(k8sClient, jobDef)
		if err != nil {
			return nil, nil, err
		}

		queriers = append(queriers, queryNS, queryCert, querySA, queryR, queryCR, queryRB, queryCRB, querySP, queryS, querySFS, queryPDB, queryES, queryJ)
		destroyers = append(destroyers, destroyNS, destroyCert, destroySA, destroyR, destroyCR, destroyRB, destroyCRB, destroySP, destroyS, destroySFS, destroyPDB, destroyES, destroyJ)

		return func() (cockroachdb.EnsureFunc, error) {
				ensurers := make([]resources.EnsureFunc, 0)
				for _, querier := range queriers {
					ensurer, err := querier()
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func() error {
					for _, ensurer := range ensurers {
						if err := ensurer(); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func() error {
				for _, destroyer := range destroyers {
					if err := destroyer(); err != nil {
						return err
					}
				}
				return nil
			},
			nil
	}
}
