package core

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

type BackupListFunc func(monitor mntr.Monitor, name string, desired *tree.Tree) ([]string, error)
