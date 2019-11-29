package adapter

import (
	"context"
	"fmt"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/external/model"
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
