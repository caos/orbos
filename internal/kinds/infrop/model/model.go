package model

import "github.com/caos/infrop/internal/core/logging"

type UserSpec struct {
	Verbose   bool
	Destroyed bool
}

type Config struct {
	Logger           logging.Logger
	ConfigID         string
	InfropVersion    string
	NodeagentRepoURL string
	NodeagentRepoKey string
	CurrentFile      string
	SecretsFile      string
	Masterkey        string
}

type Current struct{}
