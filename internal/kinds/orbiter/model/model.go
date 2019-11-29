package model

import "github.com/caos/orbiter/internal/core/logging"

type UserSpec struct {
	Verbose   bool
	Destroyed bool
}

type Config struct {
	Logger           logging.Logger
	ConfigID         string
	OrbiterVersion    string
	NodeagentRepoURL string
	NodeagentRepoKey string
	CurrentFile      string
	SecretsFile      string
	Masterkey        string
}

type Current struct{}
