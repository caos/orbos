package kubernetes

import (
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/tree"
	core "k8s.io/api/core/v1"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/mntr"
)

var deployErrors int

func AdaptFunc(
	orb *orb.Orb,
	orbiterCommit string,
	clusterID string,
	oneoff bool,
	deployOrbiter bool,
	destroyProviders func() (map[string]interface{}, error),
	whitelist func(whitelist []*orbiter.CIDR)) orbiter.AdaptFunc {

	return func(monitor mntr.Monitor, finishedChan chan bool, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {
		finished := false
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		if desiredTree.Common.Version != "v0" {
			migrate = true
		}

		desiredKind, err := parseDesiredV0(desiredTree, orb.Masterkey)
		if err != nil {
			return nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
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

		if err := desiredKind.validate(); err != nil {
			return nil, nil, migrate, err
		}

		initializeNecessarySecrets(desiredKind, orb.Masterkey)

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		whitelist([]*orbiter.CIDR{&desiredKind.Spec.Networking.PodCidr})

		var kc *string
		if desiredKind.Spec.Kubeconfig.Value != "" {
			kc = &desiredKind.Spec.Kubeconfig.Value
		}
		k8sClient := NewK8sClient(monitor, kc)

		if k8sClient.Available() && deployOrbiter {
			if err := EnsureCommonArtifacts(monitor, k8sClient); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Applying Common failed, awaiting next iteration")
			}
			if deployErrors > 50 {
				panic(err)
			}

			if err := EnsureConfigArtifacts(monitor, k8sClient, orb); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Applying configuration failed, awaiting next iteration")
			}
			if deployErrors > 50 {
				panic(err)
			}

			if err := EnsureOrbiterArtifacts(monitor, k8sClient, desiredKind.Spec.Versions.Orbiter); err != nil {
				deployErrors++
				monitor.WithFields(map[string]interface{}{
					"count": deployErrors,
					"error": err.Error(),
				}).Info("Deploying Orbiter failed, awaiting next iteration")
			} else {
				if oneoff {
					monitor.Info("Deployed Orbiter takes over control")
					finished = true
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

		go func() {
			finishedChan <- finished
		}()

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, providers map[string]interface{}) (orbiter.EnsureFunc, error) {
				ensureFunc, err := query(
					monitor,
					clusterID,
					desiredKind,
					current,
					providers,
					nodeAgentsCurrent,
					nodeAgentsDesired,
					k8sClient,
					orb.URL,
					orb.Repokey,
					orbiterCommit,
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
					return destroy(monitor, providers, k8sClient)
				}

				return orbiter.DestroyFuncGoroutine(destroyFunc)
			}, migrate, nil
	}
}
