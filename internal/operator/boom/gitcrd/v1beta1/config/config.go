package config

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/mntr"
)

type Config struct {
	Monitor          mntr.Monitor
	Git              *git.Client
	CrdDirectoryPath string
	CrdPath          string
}
