package config

import "github.com/caos/orbos/mntr"

type Config struct {
	Monitor          mntr.Monitor
	CrdUrl           string
	CrdDirectoryPath string
	CrdPath          string
	PrivateKey       []byte
	User             string
	Email            string
}
