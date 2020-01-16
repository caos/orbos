package v0

import (
	"github.com/mitchellh/mapstructure"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
)

func init() {
	build = func(serialized map[string]interface{}, _ *orbiter.Secrets, _ interface{}) (model.UserSpec, func(model.Config) ([]orbiter.Assembler, error)) {
		kind := struct{ Spec model.UserSpec }{}
		err := mapstructure.Decode(serialized, &kind)
		return kind.Spec, func(model.Config) ([]orbiter.Assembler, error) {
			return nil, err
		}
	}
}
