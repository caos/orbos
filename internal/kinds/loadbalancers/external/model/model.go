package model

import "github.com/caos/orbiter/internal/kinds/clusters/core/infra"

type UserSpec infra.Address

type Config struct{}

type Current struct {
	Address infra.Address
}
