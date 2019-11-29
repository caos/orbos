package adapter

import (
	"context"
	"fmt"

	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/loadbalancers/external/model"
)

func New() Builder {
	return builderFunc(func(spec model.UserSpec, _ operator.NodeAgentUpdater) (model.Config, Adapter, error) {
		return model.Config{}, adapterFunc(func(ctx context.Context, secrets *operator.Secrets, deps map[string]interface{}) (*model.Current, error) {
			return &model.Current{
				Address: fmt.Sprintf("%s:%d", spec.Host, spec.Port),
			}, nil
		}), nil
	})
}
