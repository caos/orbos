package config

import (
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Config struct {
	Monitor           mntr.Monitor
	CrdName           string
	BundleName        name.Bundle
	BaseDirectoryPath string
	Templator         name.Templator
}
