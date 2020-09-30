package kubernetes

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	core "k8s.io/api/core/v1"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
)

var deployErrors int

func AdaptFunc(
	clusterID string,
	oneoff bool,
	deployOrbiter bool,
	destroyProviders func() (map[string]interface{}, error),
	whitelist func(whitelist []*orbiter.CIDR)) orbiter.AdaptFunc {

	return func(monitor mntr.Monitor, finishedChan chan struct{}, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, configureFunc orbiter.ConfigureFunc, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		if desiredTree.Common.Version != "v0" {
			migrate = true
		}

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
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
			return nil, nil, nil, migrate, err
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		whitelist([]*orbiter.CIDR{&desiredKind.Spec.Networking.PodCidr})

		var kc *string
		if desiredKind.Spec.Kubeconfig != nil && desiredKind.Spec.Kubeconfig.Value != "" {
			kc = &desiredKind.Spec.Kubeconfig.Value
		}
		k8sClient := kubernetes.NewK8sClient(monitor, kc)

		if k8sClient.Available() && deployOrbiter {
			if err := kubernetes.EnsureCommonArtifacts(monitor, k8sClient); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Applying Common failed, awaiting next iteration")
			}
			if deployErrors > 50 {
				panic(err)
			}

			if err := kubernetes.EnsureOrbiterArtifacts(monitor, k8sClient, desiredKind.Spec.Versions.Orbiter); err != nil {
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

		current := &CurrentCluster{}
		currentTree.Parsed = &Current{
			Common: tree.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
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
					oneoff)
				return ensureFunc, errors.Wrapf(err, "querying %s failed", desiredKind.Common.Kind)
			}, func() error {
				defer func() {
					err = errors.Wrapf(err, "destroying %s failed", desiredKind.Common.Kind)
				}()

				providers, err := destroyProviders()
				if err != nil {
					return err
				}

				desiredKind.Spec.Kubeconfig = nil

				destroyFunc := func() error {
					return destroy(providers, k8sClient)
				}

				return orbiter.DestroyFuncGoroutine(destroyFunc)
			}, orbiter.NoopConfigure, migrate, nil
	}
}
