package kubernetes

import (
	"os"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(
	orb *orbiter.Orb,
	orbiterCommit string,
	id string,
	oneoff bool,
	deployOrbiterAndBoom bool,
	ensureProviders func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (map[string]interface{}, error),
	destroyProviders func() (map[string]interface{}, error)) orbiter.AdaptFunc {

	var deployErrors int
	return func(logger logging.Logger, desiredTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, destroyFunc orbiter.DestroyFunc, secrets map[string]*orbiter.Secret, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		if desiredTree.Common.Version != "v0" {
			migrate = true
		}

		desiredKind := &DesiredV0{
			Common: *desiredTree.Common,
			Spec:   Spec{Kubeconfig: &orbiter.Secret{Masterkey: orb.Masterkey}},
		}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if err := desiredKind.validate(); err != nil {
			return nil, nil, nil, migrate, err
		}

		if desiredKind.Spec.Kubeconfig == nil {
			desiredKind.Spec.Kubeconfig = &orbiter.Secret{Masterkey: orb.Masterkey}
		}

		if desiredKind.Spec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}

		var kc *string
		if desiredKind.Spec.Kubeconfig.Value != "" {
			kc = &desiredKind.Spec.Kubeconfig.Value
		}
		k8sClient := NewK8sClient(logger, kc)

		if k8sClient.Available() && deployOrbiterAndBoom {
			if err := ensureArtifacts(logger, k8sClient, orb, desiredKind.Spec.Versions.Orbiter, desiredKind.Spec.Versions.Boom); err != nil {
				deployErrors++
				logger.WithFields(map[string]interface{}{
					"count": deployErrors,
					"err":   err.Error(),
				}).Info(false, "Deploying Orbiter failed, awaiting next iteration")
				if deployErrors > 50 {
					panic(err)
				}
			} else {
				if oneoff {
					logger.Info(false, "Deployed Orbiter takes over control")
					os.Exit(0)
				}
				deployErrors = 0
			}
		}

		current := &CurrentCluster{}
		currentTree.Parsed = &Current{
			Common: orbiter.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
			Current: *current,
		}

		return func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error) {
				defer func() {
					err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
				}()

				providers, err := ensureProviders(psf, nodeAgentsCurrent, nodeAgentsDesired)
				if err != nil {
					return err
				}
				return ensure(
					logger,
					*desiredKind,
					current,
					providers,
					nodeAgentsCurrent,
					nodeAgentsDesired,
					psf,
					desiredKind.Spec.Kubeconfig,
					orb.URL,
					orb.Repokey,
					orbiterCommit,
					oneoff)
			}, func() error {
				defer func() {
					err = errors.Wrapf(err, "destroying %s failed", desiredKind.Common.Kind)
				}()

				providers, err := destroyProviders()
				if err != nil {
					return err
				}

				desiredKind.Spec.Kubeconfig = nil

				return destroy(logger, providers, k8sClient)
			}, map[string]*orbiter.Secret{
				"kubeconfig": desiredKind.Spec.Kubeconfig,
			}, migrate, nil
	}
}
