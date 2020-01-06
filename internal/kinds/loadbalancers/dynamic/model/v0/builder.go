package v0

import (
	"github.com/mitchellh/mapstructure"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/model"
)

func init() {
	build = func(desired map[string]interface{}, _ *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config, map[string]map[string]interface{}) (map[string]operator.Assembler, error)) {
		spec := model.UserSpec{}
		err := mapstructure.Decode(desired, &spec)

		return spec, func(cfg model.Config, deps map[string]map[string]interface{}) (map[string]operator.Assembler, error) {
			return nil, err
		}
	}
}
