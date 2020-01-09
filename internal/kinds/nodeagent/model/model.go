package model

import (
	"github.com/caos/orbiter/internal/core/operator"
)

type UserSpec struct {
	operator.NodeAgentSpec `mapstructure:",squash"`
	Verbose                bool
}

type Config struct{}

var CurrentVersion = "v0"

type Current operator.NodeAgentCurrent
