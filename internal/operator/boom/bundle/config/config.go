package config

import (
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/mntr"
)

type Config struct {
	Monitor           mntr.Monitor
	Orb               string
	CrdName           string
	BundleName        name.Bundle
	BaseDirectoryPath string
	Templator         name.Templator
}
