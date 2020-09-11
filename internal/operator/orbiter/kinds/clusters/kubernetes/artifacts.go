package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	"github.com/caos/orbos/internal/orb"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"

	"gopkg.in/yaml.v2"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caos/orbos/mntr"
)

func EnsureCommonArtifacts(monitor mntr.Monitor, client *Client) error {

	monitor.Debug("Ensuring common artifacts")

	return client.ApplyNamespace(&core.Namespace{
		ObjectMeta: mach.ObjectMeta{
			Name: "caos-system",
			Labels: map[string]string{
				"name": "caos-system",
			},
		},
	})
}

func EnsureConfigArtifacts(monitor mntr.Monitor, client *Client, orb *orb.Orb) error {
	monitor.Debug("Ensuring configuration artifacts")

	orbfile, err := yaml.Marshal(orb)
	if err != nil {
		return err
	}

	if err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "caos",
			Namespace: "caos-system",
		},
		StringData: map[string]string{
			"orbconfig": string(orbfile),
		},
	}); err != nil {
		return err
	}

	return nil
}

func EnsureZitadelArtifacts(
	monitor mntr.Monitor,
	client *Client,
	version string,
	nodeselector map[string]string,
	tolerations []core.Toleration) error {

	monitor.WithFields(map[string]interface{}{
		"zitadel": version,
	}).Debug("Ensuring zitadel artifacts")

	if version == "" {
		return nil
	}

	if err := client.ApplyServiceAccount(&core.ServiceAccount{
		ObjectMeta: mach.ObjectMeta{
			Name:      "zitadel",
			Namespace: "caos-system",
		},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRole(&rbac.ClusterRole{
		ObjectMeta: mach.ObjectMeta{
			Name: "zitadel-clusterrole",
			Labels: map[string]string{
				"app.kubernetes.io/instance":  "zitadel",
				"app.kubernetes.io/part-of":   "orbos",
				"app.kubernetes.io/component": "zitadel",
			},
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRoleBinding(&rbac.ClusterRoleBinding{
		ObjectMeta: mach.ObjectMeta{
			Name: "zitadel-clusterrolebinding",
			Labels: map[string]string{
				"app.kubernetes.io/instance":  "zitadel",
				"app.kubernetes.io/part-of":   "orbos",
				"app.kubernetes.io/component": "zitadel",
			},
		},

		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "zitadel-clusterrole",
		},
		Subjects: []rbac.Subject{{
			Kind:      "ServiceAccount",
			Name:      "zitadel",
			Namespace: "caos-system",
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyDeployment(&apps.Deployment{
		ObjectMeta: mach.ObjectMeta{
			Name:      "zitadel-operator",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "zitadel",
				"app.kubernetes.io/part-of":    "orbos",
				"app.kubernetes.io/component":  "zitadel",
				"app.kubernetes.io/managed-by": "zitadel.caos.ch",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &mach.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/instance":  "zitadel",
					"app.kubernetes.io/part-of":   "orbos",
					"app.kubernetes.io/component": "zitadel",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: mach.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/instance":  "zitadel",
						"app.kubernetes.io/part-of":   "orbos",
						"app.kubernetes.io/component": "zitadel",
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: "zitadel",
					Containers: []core.Container{{
						Name:            "zitadel",
						ImagePullPolicy: core.PullIfNotPresent,
						Image:           fmt.Sprintf("ghcr.io/caos/orbos:%s", version),
						Command:         []string{"/orbctl", "takeoff", "zitadel", "-f", "/secrets/orbconfig"},
						Args:            []string{},
						Ports: []core.ContainerPort{{
							Name:          "metrics",
							ContainerPort: 2112,
							Protocol:      "TCP",
						}},
						VolumeMounts: []core.VolumeMount{{
							Name:      "orbconfig",
							ReadOnly:  true,
							MountPath: "/secrets",
						}},
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								"cpu":    resource.MustParse("500m"),
								"memory": resource.MustParse("500Mi"),
							},
							Requests: core.ResourceList{
								"cpu":    resource.MustParse("250m"),
								"memory": resource.MustParse("250Mi"),
							},
						},
					}},
					NodeSelector: nodeselector,
					Tolerations:  tolerations,
					Volumes: []core.Volume{{
						Name: "orbconfig",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "caos",
							},
						},
					}},
					TerminationGracePeriodSeconds: int64Ptr(10),
				},
			},
		},
	}); err != nil {
		return err
	}
	monitor.WithFields(map[string]interface{}{
		"version": version,
	}).Debug("Zitadel Operator deployment ensured")

	return nil
}

func ScaleZitadelOperator(
	monitor mntr.Monitor,
	client *Client,
	replicaCount int,
) error {
	monitor.Debug("Scaling zitadel-operator")
	return client.ScaleDeployment("caos-system", "zitadel-operator", replicaCount)
}

func EnsureBoomArtifacts(monitor mntr.Monitor, client *Client, version string, tolerations k8s.Tolerations, nodeselector map[string]string, resources *k8s.Resources) error {

	monitor.WithFields(map[string]interface{}{
		"boom": version,
	}).Debug("Ensuring boom artifacts")

	if version == "" {
		return nil
	}

	if err := client.ApplyServiceAccount(&core.ServiceAccount{
		ObjectMeta: mach.ObjectMeta{
			Name:      "boom",
			Namespace: "caos-system",
		},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRole(&rbac.ClusterRole{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-clusterrole",
			Labels: map[string]string{
				"app.kubernetes.io/instance":  "boom",
				"app.kubernetes.io/part-of":   "orbos",
				"app.kubernetes.io/component": "boom",
			},
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRoleBinding(&rbac.ClusterRoleBinding{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-clusterrolebinding",
			Labels: map[string]string{
				"app.kubernetes.io/instance":  "boom",
				"app.kubernetes.io/part-of":   "orbos",
				"app.kubernetes.io/component": "boom",
			},
		},

		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "boom-clusterrole",
		},
		Subjects: []rbac.Subject{{
			Kind:      "ServiceAccount",
			Name:      "boom",
			Namespace: "caos-system",
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyDeployment(&apps.Deployment{
		ObjectMeta: mach.ObjectMeta{
			Name:      "boom",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "boom",
				"app.kubernetes.io/part-of":    "orbos",
				"app.kubernetes.io/component":  "boom",
				"app.kubernetes.io/managed-by": "boom.caos.ch",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &mach.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/instance":  "boom",
					"app.kubernetes.io/part-of":   "orbos",
					"app.kubernetes.io/component": "boom",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: mach.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/instance":  "boom",
						"app.kubernetes.io/part-of":   "orbos",
						"app.kubernetes.io/component": "boom",
					},
				},
				Spec: core.PodSpec{
					ServiceAccountName: "boom",
					Containers: []core.Container{{
						Name:            "boom",
						ImagePullPolicy: core.PullIfNotPresent,
						Image:           fmt.Sprintf("ghcr.io/caos/orbos:%s", version),
						Command:         []string{"/orbctl", "takeoff", "boom", "-f", "/secrets/orbconfig"},
						Args:            []string{},
						Ports: []core.ContainerPort{{
							Name:          "metrics",
							ContainerPort: 2112,
							Protocol:      "TCP",
						}},
						VolumeMounts: []core.VolumeMount{{
							Name:      "orbconfig",
							ReadOnly:  true,
							MountPath: "/secrets",
						}},
						Resources: core.ResourceRequirements(*resources),
					}},
					NodeSelector: nodeselector,
					Tolerations:  tolerations,
					Volumes: []core.Volume{{
						Name: "orbconfig",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "caos",
							},
						},
					}},
					TerminationGracePeriodSeconds: int64Ptr(10),
				},
			},
		},
	}); err != nil {
		return err
	}
	monitor.WithFields(map[string]interface{}{
		"version": version,
	}).Debug("Boom deployment ensured")

	if err := client.ApplyService(&core.Service{
		ObjectMeta: mach.ObjectMeta{
			Name:      "boom",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "boom",
				"app.kubernetes.io/part-of":    "orbos",
				"app.kubernetes.io/component":  "boom",
				"app.kubernetes.io/managed-by": "boom.caos.ch",
			},
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{{
				Name:       "metrics",
				Protocol:   "TCP",
				Port:       2112,
				TargetPort: intstr.FromInt(2112),
			}},
			Selector: map[string]string{
				"app.kubernetes.io/instance":  "boom",
				"app.kubernetes.io/part-of":   "orbos",
				"app.kubernetes.io/component": "boom",
			},
			Type: core.ServiceTypeClusterIP,
		},
	}); err != nil {
		return err
	}
	monitor.Debug("Boom service ensured")

	return nil
}

func EnsureOrbiterArtifacts(monitor mntr.Monitor, client *Client, orbiterversion string) error {
	monitor.WithFields(map[string]interface{}{
		"orbiter": orbiterversion,
	}).Debug("Ensuring orbiter artifacts")

	if orbiterversion == "" {
		return nil
	}

	if err := client.ApplyDeployment(&apps.Deployment{
		ObjectMeta: mach.ObjectMeta{
			Name:      "orbiter",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "orbiter",
				"app.kubernetes.io/part-of":    "orbos",
				"app.kubernetes.io/component":  "orbiter",
				"app.kubernetes.io/managed-by": "orbiter.caos.ch",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &mach.LabelSelector{
				MatchLabels: map[string]string{
					"app": "orbiter",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: mach.ObjectMeta{
					Labels: map[string]string{
						"app": "orbiter",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Name:            "orbiter",
						ImagePullPolicy: core.PullIfNotPresent,
						Image:           "ghcr.io/caos/orbos:" + orbiterversion,
						Command:         []string{"/orbctl", "--orbconfig", "/etc/orbiter/orbconfig", "takeoff", "orbiter", "--recur", "--ingestion="},
						VolumeMounts: []core.VolumeMount{{
							Name:      "keys",
							ReadOnly:  true,
							MountPath: "/etc/orbiter",
						}},
						Ports: []core.ContainerPort{{
							Name:          "metrics",
							ContainerPort: 9000,
						}},
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								"cpu":    resource.MustParse("500m"),
								"memory": resource.MustParse("500Mi"),
							},
							Requests: core.ResourceList{
								"cpu":    resource.MustParse("250m"),
								"memory": resource.MustParse("250Mi"),
							},
						},
					}},
					Volumes: []core.Volume{{
						Name: "keys",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "caos",
								Optional:   boolPtr(false),
							},
						},
					}},
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					Tolerations: []core.Toleration{{
						Key:      "node-role.kubernetes.io/master",
						Effect:   "NoSchedule",
						Operator: "Exists",
					}},
				},
			},
		},
	}); err != nil {
		return err
	}
	monitor.WithFields(map[string]interface{}{
		"version": orbiterversion,
	}).Debug("Orbiter deployment ensured")

	if err := client.ApplyService(&core.Service{
		ObjectMeta: mach.ObjectMeta{
			Name:      "orbiter",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "orbiter",
				"app.kubernetes.io/part-of":    "orbos",
				"app.kubernetes.io/component":  "orbiter",
				"app.kubernetes.io/managed-by": "orbiter.caos.ch",
			},
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{{
				Name:       "metrics",
				Protocol:   "TCP",
				Port:       9000,
				TargetPort: intstr.FromInt(9000),
			}},
			Selector: map[string]string{
				"app": "orbiter",
			},
			Type: core.ServiceTypeClusterIP,
		},
	}); err != nil {
		return err
	}
	monitor.Debug("Orbiter service ensured")

	if _, err := client.set.AppsV1().Deployments("kube-system").Patch(context.Background(), "coredns", types.StrategicMergePatchType, []byte(`
{
  "spec": {
    "template": {
      "spec": {
        "affinity": {
          "podAntiAffinity": {
            "preferredDuringSchedulingIgnoredDuringExecution": [{
              "weight": 100,
              "podAffinityTerm": {
                "topologyKey": "kubernetes.io/hostname"
              }
            }]
          }
        }
      }
    }
  }
}`), mach.PatchOptions{}); err != nil {
		return err
	}

	monitor.Debug("CoreDNS deployment patched")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
