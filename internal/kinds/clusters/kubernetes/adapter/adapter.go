package adapter

import (
	"context"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/model"
)

func New(params model.Parameters) Builder {
	return builderFunc(func(spec model.UserSpec, nodeAgentUpdater orbiter.NodeAgentUpdater) (model.Config, Adapter, error) {

		cfg := model.Config{
			Spec:   spec,
			Params: params,
		}

		if cfg.Spec.Verbose && !cfg.Params.Logger.IsVerbose() {
			cfg.Params.Logger = cfg.Params.Logger.Verbose()
		}

		return cfg, adapterFunc(func(ctx context.Context, secrets *orbiter.Secrets, ensuredDependencies map[string]interface{}) (*model.Current, error) {

			poolIsConfigured := func(poolSpec *model.Pool, infra map[string]map[string]infra.Pool) error {
				prov, ok := infra[poolSpec.Provider]
				if !ok {
					return errors.Errorf("provider %s not configured", poolSpec.Provider)
				}
				if _, ok := prov[poolSpec.Pool]; !ok {
					return errors.Errorf("pool %s not configured on provider %s", poolSpec.Provider, poolSpec.Pool)
				}
				return nil
			}

			curr := &model.Current{
				Status:   "maintaining",
				Computes: make(map[string]*model.Compute),
			}

			cloudPools := make(map[string]map[string]infra.Pool)
			providersCleanupped := make([]<-chan error, 0)
			var kubeAPIAddress infra.Address

			for providerName, provider := range ensuredDependencies {
				if cloudPools[providerName] == nil {
					cloudPools[providerName] = make(map[string]infra.Pool)
				}
				prov := provider.(infra.ProviderCurrent)
				providerPools := prov.Pools()
				providerIngresses := prov.Ingresses()
				providerCleanupped := prov.Cleanupped()
				providersCleanupped = append(providersCleanupped, providerCleanupped)
				for providerPoolName, providerPool := range providerPools {
					cloudPools[providerName][providerPoolName] = providerPool
					if spec.ControlPlane.Provider == providerName && spec.ControlPlane.Pool == providerPoolName {
						kubeAPIAddress = providerIngresses["kubeapi"]
						cfg.Params.Logger.WithFields(map[string]interface{}{
							"address": kubeAPIAddress,
						}).Debug("Found kubernetes api address")
					}
				}
			}

			if err := poolIsConfigured(spec.ControlPlane, cloudPools); err != nil {
				return curr, err
			}

			for _, w := range spec.Workers {
				if err := poolIsConfigured(w, cloudPools); err != nil {
					return curr, err
				}
			}

			k8sClient := k8s.New(cfg.Params.Logger, nil)
			if err := ensureCluster(&cfg, curr, cloudPools, kubeAPIAddress, secrets, k8sClient); err != nil {
				return curr, errors.Wrap(err, "ensuring cluster failed")
			}

			if spec.Destroyed {
				return nil, infra.Destroy(ensuredDependencies)
			}

			for _, cleanupped := range providersCleanupped {
				if err := <-cleanupped; err != nil {
					return curr, err
				}
			}

			return curr, nil
		}), nil
	})
}

/*
func before(curr *model.Current, k8s *k8s.Client, selftag, repourl, repokey, masterkey string) error {

	curr.OrbiterDeployment = map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]string{
			"name":      "orbiter",
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"replicas": 1,
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{
					"name": "orbiter",
				},
			},
			"strategy": map[string]interface{}{
				"type": apps.RecreateDeploymentStrategyType,
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]string{
						"prometheus.io/port": "3031",
					},
					"labels": map[string]string{
						"name": "orbiter",
					},
				},
				"spec": map[string]interface{}{
					"nodeSelector": map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					"tolerations": []map[string]interface{}{{
						"key":      "node-role.kubernetes.io/master",
						"operator": core.TolerationOpEqual,
						"value":    "",
						"effect":   core.TaintEffectNoSchedule,
					}},
					// TODO: Remove before open sourcing #39
					"imagePullSecrets": []map[string]string{{
						"name": "orbiterregistry",
					}},
					"containers": []map[string]interface{}{{
						"name":            "orbiter",
						"imagePullPolicy": core.PullAlways,
						"image":           fmt.Sprintf("docker.pkg.github.com/caos/orbiter/orbiter:%s", selftag),
						"command":         []string{"/artifacts/orbiter", "--recur", "--repourl", repourl},
						"volumeMounts": []map[string]interface{}{{
							"name":      "keys",
							"readOnly":  true,
							"mountPath": "/etc/orbiter",
						}},
					}},
					"volumes": []map[string]interface{}{{
						"name": "keys",
						"volumeSource": map[string]interface{}{
							"secret": map[string]interface{}{
								"secretName": "caos",
								"optional":   "false",
							},
						},
					}},
				},
			},
		},
	}
	// TODO: Only initially create secret
	if err := k8s.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "caos",
			Namespace: "kube-system",
		},
		StringData: map[string]string{
			"repokey":   repokey,
			"masterkey": masterkey,
		},
	}); err != nil {
		return err
	}

	fluxProbes := &core.Probe{
		Handler: core.Handler{
			HTTPGet: &core.HTTPGetAction{
				Port: intstr.FromInt(3030),
				Path: "/api/flux/v6/identity.pub",
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
	}

	if err := k8s.ApplyDeployment(&apps.Deployment{
		ObjectMeta: mach.ObjectMeta{
			Name:      "flux-root",
			Namespace: "kube-system",
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &mach.LabelSelector{
				MatchLabels: map[string]string{
					"app": "flux-root",
				},
			},
			Strategy: apps.DeploymentStrategy{
				Type: apps.RecreateDeploymentStrategyType,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: mach.ObjectMeta{
					Labels: map[string]string{
						"name": "flux-root",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Name:            "flux-root",
						Image:           "eu.gcr.io/caos-ops/flux:gopass",
						ImagePullPolicy: core.PullAlways,
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList(map[core.ResourceName]resource.Quantity{
								core.ResourceCPU:    resource.MustParse("50m"),
								core.ResourceMemory: resource.MustParse("64Mi"),
							}),
						},
						Ports: []core.ContainerPort{{
							ContainerPort: 3030,
						}},
						LivenessProbe:  fluxProbes,
						ReadinessProbe: fluxProbes,
						Args:           []string{},
						Command:        []string{"/artifacts/orbiter", "--recur", "--repourl", repourl},
						VolumeMounts: []core.VolumeMount{{
							Name:      "git-key",
							ReadOnly:  true,
							MountPath: "/etc/fluxd/ssh",
						}},
					}},
					Volumes: []core.Volume{{
						Name: "git-key",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "caos",
								Optional:   boolPtr(false),
								Items: []core.KeyToPath{{
									Key: "repokey",
								}},
							},
						},
					}},
				},
			},
		},
	}); err != nil {
		return err
	}
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }
*/
