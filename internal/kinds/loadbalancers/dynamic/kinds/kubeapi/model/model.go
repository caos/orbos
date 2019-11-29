package model

import "github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/model"

type UserSpec struct {
	IP string
}

type Config struct{}

type Current struct {
	model.UserSpec `yaml:"-"`
}

func (c *Current) Overwrite() model.UserSpec {
	return c.UserSpec
}
