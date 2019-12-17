package model

import (
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/logging"
)

type Compute struct {
	ID         string
	Hostname   string
	DomainName string
}

type UserSpec struct {
	Verbose             bool
	RemoteUser          string
	RemotePublicKeyPath string
	Pools               map[string][]*Compute
}

type Config struct {
	Logger       logging.Logger
	ID           string
	Healthchecks string
}

type Current struct {
	infra.ProviderCurrent
}
