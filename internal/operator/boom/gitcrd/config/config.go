package config

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
)

type Config struct {
	CrdDirectoryPath string
	CrdPath          string
	Git              *git.Client
	Monitor          mntr.Monitor
}
