package api

import (
	"strings"

	"github.com/caos/orbos/internal/operator/boom/api/common"
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/migrate"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/metrics"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

const (
	boomPrefix = "caos.ch"
)

func ParseToolset(desiredTree *tree.Tree) (*latest.Toolset, bool, map[string]*secret.Secret, map[string]*secret.Existing, string, string, error) {
	desiredKindCommon := common.New()
	if err := desiredTree.Original.Decode(desiredKindCommon); err != nil {
		metrics.WrongCRDFormat()
		return nil, false, nil, nil, "", "", errors.Wrap(err, "parsing desired state failed")
	}
	if desiredKindCommon.Kind != "Boom" {
		return nil, false, nil, nil, "", "", errors.New("Kind unknown")
	}

	if !strings.HasPrefix(desiredKindCommon.APIVersion, boomPrefix) {
		metrics.UnsupportedAPIGroup()
		return nil, false, nil, nil, "", "", errors.New("Group unknown")
	}

	switch desiredKindCommon.APIVersion {
	case boomPrefix + "/v1beta1":
		old, _, err := v1beta1.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, nil, "", "", err
		}
		v1beta2Toolset, _ := migrate.V1beta1Tov1beta2(old)
		v1Toolset, secrets, existing := migrate.V1beta2Tov1(v1beta2Toolset)

		metrics.SuccessfulUnmarshalCRD()
		return v1Toolset, true, secrets, existing, desiredKindCommon.Kind, "v1beta1", err
	case boomPrefix + "/v1beta2":
		v1beta2Toolset, _, err := v1beta2.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, nil, "", "", err
		}
		v1Toolset, secrets, existing := migrate.V1beta2Tov1(v1beta2Toolset)

		metrics.SuccessfulUnmarshalCRD()
		return v1Toolset, true, secrets, existing, desiredKindCommon.Kind, "v1beta2", nil
	case boomPrefix + "/v1":
		desiredKind, secrets, existing, err := latest.ParseToolset(desiredTree)
		if err != nil {
			return nil, false, nil, nil, "", "", err
		}
		metrics.SuccessfulUnmarshalCRD()
		return desiredKind, false, secrets, existing, desiredKindCommon.Kind, "v1", nil
	default:
		metrics.UnsupportedVersion()
		return nil, false, nil, nil, "", "", errors.New("APIVersion unknown")
	}

}
