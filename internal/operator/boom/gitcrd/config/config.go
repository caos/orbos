package config

import (
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
)

type Config struct {
	CrdDirectoryPath string
	CrdPath          string
	Git              *git.Client
	Monitor          mntr.Monitor
}
