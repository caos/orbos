package adapter

import (
	"context"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/orb/model"
)

func New(cfg *model.Config) Builder {
	return builderFunc(func(spec model.UserSpec, _ operator.NodeAgentUpdater) (model.Config, Adapter, error) {

		if spec.Verbose && !cfg.Logger.IsVerbose() {
			cfg.Logger = cfg.Logger.Verbose()
		}

		return *cfg, adapterFunc(func(context.Context, *operator.Secrets, map[string]interface{}) (*model.Current, error) {
			return nil, nil
		}), nil
	})
}
