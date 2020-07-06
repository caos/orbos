package api

import (
	"github.com/caos/orbos/internal/operator/boom/api/common"
	"github.com/caos/orbos/internal/operator/boom/api/migrate"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func ParseToolset(desiredTree *tree.Tree) (*v1beta2.Toolset, bool, error) {
	desiredKindCommon := common.New()
	if err := desiredTree.Original.Decode(desiredKindCommon); err != nil {
		return nil, false, errors.Wrap(err, "parsing desired state failed")
	}

	switch desiredKindCommon.APIVersion {
	case "boom.caos.ch/v1beta1":
		old, err := v1beta1.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, err
		}
		return migrate.V1beta1Tov1beta2(old), true, err
	case "boom.caos.ch/v1beta2":
		desiredKind, err := v1beta2.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, err
		}
		return desiredKind, false, nil
	default:
		return nil, false, errors.New("APIVersion unknown")
	}

}

func SecretsFunc() secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		desiredKindCommon := common.New()
		if err := desiredTree.Original.Decode(desiredKindCommon); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}

		switch desiredKindCommon.APIVersion {
		case "boom.caos.ch/v1beta1":
			return v1beta1.SecretsFunc(desiredTree)
		case "boom.caos.ch/v1beta2":
			return v1beta2.SecretsFunc(desiredTree)
		default:
			return nil, errors.New("APIVersion unknown")
		}
	}
}
