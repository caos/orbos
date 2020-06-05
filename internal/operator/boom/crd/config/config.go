package config

import (
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Config struct {
	Monitor mntr.Monitor
	Version name.Version
}
