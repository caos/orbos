package adapter

import (
	"context"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/external/model"
)

func New() Builder {
	return builderFunc(func(spec model.UserSpec, _ orbiter.NodeAgentUpdater) (model.Config, Adapter, error) {
		return model.Config{}, adapterFunc(func(ctx context.Context, secrets *orbiter.Secrets, deps map[string]interface{}) (*model.Current, error) {
			return &model.Current{
				Address: infra.Address(spec),
			}, nil
		}), nil
	})
}
