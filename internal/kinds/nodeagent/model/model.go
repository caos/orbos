package model

import (
	"github.com/caos/infrop/internal/core/operator"
)

type UserSpec struct {
	operator.NodeAgentSpec `mapstructure:",squash"`
	Verbose                bool
}

type Config struct{}

type Current operator.NodeAgentCurrent
