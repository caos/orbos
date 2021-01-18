package api

import (
	"github.com/caos/orbos/internal/operator/boom/api/common"
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/migrate"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func ParseToolset(desiredTree *tree.Tree) (*latest.Toolset, bool, map[string]*secret.Secret, string, string, error) {
	desiredKindCommon := common.New()
	if err := desiredTree.Original.Decode(desiredKindCommon); err != nil {
		return nil, false, nil, "", "", errors.Wrap(err, "parsing desired state failed")
	}

	if desiredKindCommon.Kind != "Boom" {
		return nil, false, nil, "", "", errors.New("Kind unknown")
	}

	switch desiredKindCommon.APIVersion {
	case "caos.ch/v1beta1":
		old, _, err := v1beta1.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, "", "", err
		}
		v1beta2Toolset, _ := migrate.V1beta1Tov1beta2(old)
		v1Toolset, secrets := migrate.V1beta2Tov1(v1beta2Toolset)

		return v1Toolset, true, secrets, desiredKindCommon.Kind, "v1beta1", err
	case "caos.ch/v1beta2":
		v1beta2Toolset, _, err := v1beta2.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, "", "", err
		}
		v1Toolset, secrets := migrate.V1beta2Tov1(v1beta2Toolset)

		return v1Toolset, true, secrets, desiredKindCommon.Kind, "v1beta2", nil
	case "caos.ch/v1":
		desiredKind, secrets, err := latest.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, "", "", err
		}
		return desiredKind, false, secrets, desiredKindCommon.Kind, "v1", nil
	default:
		return nil, false, nil, "", "", errors.New("APIVersion unknown")
	}

}
