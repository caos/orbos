package kubernetes

import (
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

var deployErrors int

func AdaptFunc(
	apiLabels *labels.API,
	clusterID string,
	oneoff bool,
	deployOrbiter bool,
	pprof bool,
	destroyProviders func(map[string]interface{}) (map[string]interface{}, error),
	whitelist func(whitelist []*orbiter.CIDR),
	gitClient *git.Client,
) orbiter.AdaptFunc {

	return func(
		monitor mntr.Monitor,
		finishedChan chan struct{},
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		queryFunc orbiter.QueryFunc,
		destroyFunc orbiter.DestroyFunc,
		configureFunc orbiter.ConfigureFunc,
		migrate bool,
		secrets map[string]*secret.Secret,
		err error,
	) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("building %s failed: %w", desiredTree.Common.Kind, err)
			}
		}()

		if desiredTree.Common.Version() != "v0" {
			migrate = true
		}

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, nil, migrate, nil, fmt.Errorf("parsing desired state failed: %w", err)
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.ControlPlane.Taints == nil {
			taints := Taints([]Taint{{
				Key:    "node-role.kubernetes.io/master",
				Effect: core.TaintEffectNoSchedule,
			}})
			desiredKind.Spec.ControlPlane.Taints = &taints
			migrate = true
		}

		for _, workers := range desiredKind.Spec.Workers {
			if workers.Taints == nil {
				taints := Taints(make([]Taint, 0))
				workers.Taints = &taints
				migrate = true
			}
		}

		if err := desiredKind.validate(); err != nil {
			return nil, nil, nil, migrate, nil, err
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		whitelist([]*orbiter.CIDR{&desiredKind.Spec.Networking.PodCidr})

		k8sClient := tryToConnect(monitor, *desiredKind)

		if k8sClient != nil && deployOrbiter {
			if err := kubernetes.EnsureCaosSystemNamespace(monitor, k8sClient); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Applying Common failed, awaiting next iteration")
			}
			if deployErrors > 50 {
				panic(err)
			}

			imageRegistry := desiredKind.Spec.CustomImageRegistry
			if imageRegistry == "" {
				imageRegistry = "ghcr.io"
			}

			if err := kubernetes.EnsureOrbiterArtifacts(
				monitor,
				apiLabels,
				k8sClient,
				pprof,
				desiredKind.Spec.Versions.Orbiter,
				imageRegistry,
			); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Deploying Orbiter failed, awaiting next iteration")
			} else {
				if oneoff {
					monitor.Info("Deployed Orbiter takes over control")
					finishedChan <- struct{}{}
				}
				deployErrors = 0
			}
			if deployErrors > 50 {
				panic(err)
			}
		}

		currentKind := "orbiter.caos.ch/KubernetesCluster"
		current := &CurrentCluster{}
		currentTree.Parsed = &Current{
			Common:  *(tree.NewCommon(currentKind, "v0", false)),
			Current: current,
		}

		return func(nodeAgentsCurrent *common.CurrentNodeAgents, nodeAgentsDesired *common.DesiredNodeAgents, providers map[string]interface{}) (orbiter.EnsureFunc, error) {
				ensureFunc, err := query(
					monitor,
					clusterID,
					desiredKind,
					current,
					providers,
					nodeAgentsCurrent,
					nodeAgentsDesired,
					k8sClient,
					oneoff,
					gitClient,
				)
				if err != nil {
					err = fmt.Errorf("querying %s failed: %w", desiredKind.Common.Kind, err)
				}
				return ensureFunc, err
			}, func(delegate map[string]interface{}) error {
				defer func() {
					if err != nil {
						err = fmt.Errorf("destroying %s failed: %w", desiredKind.Common.Kind, err)
					}
				}()

				if k8sClient != nil {
					volumes, err := k8sClient.ListPersistentVolumes()
					if err != nil {
						return err
					}

					volumeNames := make([]infra.Volume, len(volumes.Items))
					for idx := range volumes.Items {
						volumeNames[idx] = infra.Volume{Name: volumes.Items[idx].Name}
					}
					delegate[currentKind] = volumeNames
				}

				providers, err := destroyProviders(delegate)
				if err != nil {
					return err
				}

				desiredKind.Spec.Kubeconfig = nil

				return destroy(providers, k8sClient)
			},
			orbiter.NoopConfigure,
			migrate,
			getSecretsMap(desiredKind),
			nil
	}
}
