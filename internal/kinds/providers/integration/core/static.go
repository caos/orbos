// +build test integration

package core

import (
	"context"
	"os"

	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/kinds/providers/core"
	"github.com/caos/infrop/internal/kinds/providers/static"
	staticconfig "github.com/caos/infrop/internal/kinds/providers/static/config"
	logcontext "github.com/caos/infrop/internal/edge/logger/context"
	"github.com/caos/infrop/internal/edge/logger/stdlib"

	"github.com/spf13/viper"
)

type staticProvider struct {
	config  *viper.Viper
	secrets *viper.Viper
}

func Static(config *viper.Viper, secrets *viper.Viper) Provider {
	return &staticProvider{config, secrets}
}

func (s *staticProvider) Assemble(operatorID string, configuredPools []string, configuredLoadBalancers []*LoadBalancer) (infra.Provider, core.ComputesService, interface{}, error) {

	ctx := context.Background()
	assembly, err := staticconfig.New(ctx, s.config, map[string]interface{}{
		"operatorId": operatorID,
	}, s.secrets).Assemble()
	if err != nil {
		return nil, nil, nil, err
	}

	logger := logcontext.Add(stdlib.New(os.Stdout)).Verbose()
	computesSvc := static.NewComputesService(logger, assembly)

	return static.New(logger, assembly), computesSvc, assembly, nil
}
