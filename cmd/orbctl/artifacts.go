package main

import (
	"os"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureArtifacts(logger logging.Logger, secrets *operator.Secrets, secretsNamespace, repourl, repokey, masterkey, orbiterversion string, boomversion string) error {

	kc, err := secrets.Read(secretsNamespace + "_kubeconfig")
	if err != nil {
		return nil
	}

	kcStr := string(kc)
	client := k8s.New(logger, &kcStr)

	if _, err := client.ApplyNamespace(&core.Namespace{
		ObjectMeta: mach.ObjectMeta{
			Name: "caos-system",
			Labels: map[string]string{
				"name": "caos-system",
			},
		},
	}); err != nil {
		return err
	}

	if _, err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "caos",
			Namespace: "caos-system",
		},
		StringData: map[string]string{
			"repokey":   repokey,
			"masterkey": masterkey,
		},
	}); err != nil {
		return err
	}

	if _, err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "public-github-packages",
			Namespace: "caos-system",
		},
		Type: core.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			core.DockerConfigJsonKey: `{
		"auths": {
				"docker.pkg.github.com": {
						"auth": "aW1ncHVsbGVyOmU2NTAxMWI3NDk1OGMzOGIzMzcwYzM5Zjg5MDlkNDE5OGEzODBkMmM="
				}
		}
}`,
		},
	}); err != nil {
		return err
	}

	if orbiterversion != "" {
		created, err := client.ApplyDeployment(&apps.Deployment{
			ObjectMeta: mach.ObjectMeta{
				Name:      "orbiter",
				Namespace: "caos-system",
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
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/master": "",
						},
						Tolerations: []core.Toleration{{
							Key:      "node-role.kubernetes.io/master",
							Operator: "Equal",
							Value:    "",
							Effect:   "NoSchedule",
						}},
						ImagePullSecrets: []core.LocalObjectReference{{
							Name: "public-github-packages",
						}},
						Containers: []core.Container{{
							Name:            "orbiter",
							ImagePullPolicy: core.PullIfNotPresent,
							Image:           "docker.pkg.github.com/caos/orbiter/orbiter:" + orbiterversion,
							Command:         []string{"/orbctl", "--repourl", repourl, "--repokey-file", "/etc/orbiter/repokey", "--masterkey-file", "/etc/orbiter/masterkey", "takeoff", "--recur"},
							VolumeMounts: []core.VolumeMount{{
								Name:      "keys",
								ReadOnly:  true,
								MountPath: "/etc/orbiter",
							}},
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
					},
				},
			},
		})
		if err != nil {
			return err
		}

		if created {
			os.Exit(0)
		}
	}

	return nil
	/*
		if boomversion == "" {
			return nil
		}

		_, err = client.ApplyDeployment(&apps.Deployment{
			ObjectMeta: mach.ObjectMeta{
				Name:      "boom",
				Namespace: "caos-system",
			},
			Spec: apps.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &mach.LabelSelector{
					MatchLabels: map[string]string{
						"app": "boom",
					},
				},
				Template: core.PodTemplateSpec{
					ObjectMeta: mach.ObjectMeta{
						Labels: map[string]string{
							"app": "boom",
						},
					},
					Spec: core.PodSpec{
						ImagePullSecrets: []core.LocalObjectReference{{
							Name: "public-github-packages",
						}},
						Containers: []core.Container{{
							Name:            "boom",
							ImagePullPolicy: core.PullIfNotPresent,
							Image:           fmt.Sprintf("docker.pkg.github.com/caos/boom/boom:%s", boomversion),
							Command:         []string{"/boom"},
							Args: []string{
								"--metrics-addr", "127.0.0.1:8080",
								"--enable-leader-election",
								// "--git-crd-secret", "/secrets/tools-secret/id_rsa-toolsop-tools-read",
							},
						}},
					},
				},
			},
		})
		return err
	*/
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }
