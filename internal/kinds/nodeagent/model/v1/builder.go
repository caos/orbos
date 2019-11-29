package v1

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/nodeagent/model"
)

func init() {
	build = func(desired map[string]interface{}, _ *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config, map[string]map[string]interface{}) (map[string]operator.Assembler, error)) {

		userSpec := model.UserSpec{}
		err := mapstructure.Decode(desired, &userSpec)

		return userSpec, func(_ model.Config, deps map[string]map[string]interface{}) (map[string]operator.Assembler, error) {

			if err != nil {
				return nil, err
			}

			if len(deps) > 0 {
				return nil, errors.New("Node agent does not take dependencies")
			}
			return nil, nil
		}
	}
}
