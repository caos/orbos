package legacycf

import (
	"context"
	"errors"
	"fmt"

	opcore "github.com/caos/orbos/v5/internal/operator/core"
	"github.com/caos/orbos/v5/internal/operator/networking/kinds/networking/core"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/labels"
	"github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
	"github.com/caos/orbos/v5/pkg/treelabels"
)

func AdaptFunc(
	ctx context.Context,
	namespace string,
	operatorLabels *labels.Operator,
) opcore.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		opcore.QueryFunc,
		opcore.DestroyFunc,
		map[string]*secret.Secret,
		map[string]*secret.Existing,
		bool,
		error,
	) {
		internalMonitor := monitor.WithField("kind", "legacycf")
		apiLabels := treelabels.MustForAPI(desiredTree, operatorLabels)

		desiredKind, err := parseDesired(desiredTree)
		if err != nil {
			return nil, nil, nil, nil, false, fmt.Errorf("parsing desired state failed: %w", err)
		}
		desiredTree.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		if desiredKind.Spec == nil {
			return nil, nil, nil, nil, false, errors.New("No specs found")
		}

		if err := desiredKind.Spec.Validate(); err != nil {
			return nil, nil, nil, nil, false, err
		}

		internalSpec, current := desiredKind.Spec.Internal(namespace, apiLabels)

		legacyQuerier, legacyDestroyer, readyCertificate, err := adaptFunc(ctx, monitor, internalSpec)
		if err != nil {
			return nil, nil, nil, nil, false, err
		}
		current.ReadyCertificate = readyCertificate

		queriers := []opcore.QueryFunc{
			legacyQuerier,
		}
		currentTree.Parsed = current

		secrets, existing := getSecretsMap(desiredKind)

		return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (opcore.EnsureFunc, error) {
				if err := desiredKind.Spec.ValidateSecrets(); err != nil {
					return nil, err
				}

				core.SetQueriedForNetworking(queried, currentTree)
				internalMonitor.Info("set current state legacycf")

				return opcore.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
			},
			opcore.DestroyersToDestroyFunc(internalMonitor, []opcore.DestroyFunc{legacyDestroyer}),
			secrets,
			existing,
			false,
			nil
	}
}
