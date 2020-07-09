package config

import "github.com/caos/orbos/mntr"

type Config struct {
	Monitor          mntr.Monitor
	CrdDirectoryPath string
	CrdPath          string
	User             string
	Email            string
}
