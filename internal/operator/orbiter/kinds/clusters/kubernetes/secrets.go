package kubernetes

import (
	"github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/mntr"
)

func SecretFunc(orb *orb.Orb) secret.Func {

	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{
			Common: *desiredTree.Common,
			Spec:   Spec{Kubeconfig: &secret.Secret{Masterkey: orb.Masterkey}},
		}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if err := desiredKind.validate(); err != nil {
			return nil, err
		}

		if desiredKind.Spec.Kubeconfig == nil {
			desiredKind.Spec.Kubeconfig = &secret.Secret{Masterkey: orb.Masterkey}
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		current := &CurrentCluster{}
		currentTree.Parsed = &Current{
			Common: tree.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
			Current: current,
		}

		return map[string]*secret.Secret{
			"kubeconfig": desiredKind.Spec.Kubeconfig,
		}, nil
	}
}
