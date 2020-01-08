package v0

import (
	"github.com/mitchellh/mapstructure"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/external/model"
)

func init() {
	build = func(serialized map[string]interface{}, _ *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config) ([]operator.Assembler, error)) {
		kind := struct{ Spec model.UserSpec }{}
		err := mapstructure.Decode(serialized, &kind)
		return kind.Spec, func(model.Config) ([]operator.Assembler, error) {
			return nil, err
		}
	}
}
