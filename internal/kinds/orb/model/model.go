package model

import "github.com/caos/orbiter/logging"

type UserSpec struct {
	Verbose   bool
	Destroyed bool
}

type Config struct {
	Logger             logging.Logger
	ConfigID           string
	OrbiterCommit      string
	NodeagentRepoURL   string
	NodeagentRepoKey   string
	CurrentFile        string
	SecretsFile        string
	Masterkey          string
	ConnectFromOutside bool
}

var CurrentVersion = "v0"

type Current struct{}
