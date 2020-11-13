package config

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/mntr"
)

type Config struct {
	CrdDirectoryPath string
	CrdPath          string
	Git              *git.Client
	Monitor          mntr.Monitor
	Deploy           bool
}
