package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(
	logger logging.Logger,
	orb *orbiter.Orb,
	orbiterCommit string,
	id string,
	takeoff bool,
	deployOrbiterAndBoom bool,
	ensureProviders func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (map[string]interface{}, error),
	destroyProviders func() (map[string]interface{}, error)) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, destroyFunc orbiter.DestroyFunc, readSecretFunc orbiter.ReadSecretFunc, writeSecretFunc orbiter.WriteSecretFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: *desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		secretsKind := &SecretsV0{
			Common:  *secretsTree.Common,
			Secrets: Secrets{Kubeconfig: &orbiter.Secret{Masterkey: orb.Masterkey}},
		}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		if secretsKind.Secrets.Kubeconfig == nil {
			secretsKind.Secrets.Kubeconfig = &orbiter.Secret{Masterkey: orb.Masterkey}
		}

		if deployOrbiterAndBoom && secretsKind.Secrets.Kubeconfig.Value != "" {
			if err := ensureArtifacts(logger, secretsKind.Secrets.Kubeconfig, orb, takeoff, desiredKind.Spec.Versions.Orbiter, desiredKind.Spec.Versions.Boom); err != nil {
				return nil, nil, nil, nil, err
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

		secretsMap := map[string]*orbiter.Secret{
			"kubeconfig": secretsKind.Secrets.Kubeconfig,
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
					secretsKind.Secrets.Kubeconfig,
					orb.URL,
					orb.Repokey,
					orbiterCommit)
			}, func() error {
				defer func() {
					err = errors.Wrapf(err, "destroying %s failed", desiredKind.Common.Kind)
				}()

				providers, err := destroyProviders()
				if err != nil {
					return err
				}

				return destroy(providers, secretsKind.Secrets.Kubeconfig)
			}, func(path []string) (string, error) {
				return orbiter.AdaptReadSecret(path, nil, secretsMap)
			}, func(path []string, value string) error {
				return orbiter.AdaptWriteSecret(path, value, nil, secretsMap)
			}, nil
	}
}
