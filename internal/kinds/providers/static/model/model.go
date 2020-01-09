package model

import (
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/logging"
)

type Compute struct {
	ID       string
	Hostname string
	IP       string
}

type UserSpec struct {
	Verbose             bool
	RemoteUser          string
	RemotePublicKeyPath string
	Pools               map[string][]*Compute
	Hoster              string
}

type Config struct {
	Logger       logging.Logger
	ID           string
	Healthchecks string
}

var CurrentVersion = "v0"

type Current struct {
	infra.ProviderCurrent
}
