// Code generated by "gen-kindstubs -parentpath=github.com/caos/orbos/internal/operator/orbiter/kinds/providers -versions=v0 -kind=orbiter.caos.ch/EC2Provider from file gen.go"; DO NOT EDIT.

package v0

import (
	"errors"

	"github.com/caos/orbos/internal/core/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ec2/model"
)

var build func(map[string]interface{}, *orbiter.Secrets, interface{}) (model.UserSpec, func(model.Config) ([]orbiter.Assembler, error))

func Build(spec map[string]interface{}, secrets *orbiter.Secrets, dependant interface{}) (model.UserSpec, func(cfg model.Config) ([]orbiter.Assembler, error)) {
	if build != nil {
		return build(spec, secrets, dependant)
	}
	return model.UserSpec{}, func(_ model.Config) ([]orbiter.Assembler, error) {
		return nil, errors.New("Version v0 for kind orbiter.caos.ch/EC2Provider is not yet supported")
	}
}
