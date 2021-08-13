package main

import (
	"fmt"
	"os"

	"github.com/caos/orbos/mntr"
)

var (
	// Build arguments
	gitCommit          = "none"
	version            = "none"
	githubClientID     = "none"
	githubClientSecret = "none"
	monitor            = mntr.Monitor{
		OnInfo:         mntr.LogMessage,
		OnChange:       mntr.LogMessage,
		OnError:        mntr.LogError,
		OnRecoverPanic: mntr.LogPanic,
	}
)

func main() {

	defer func() { monitor.RecoverPanic(recover()) }()

	rootCmd, getRootValues := RootCommand()
	rootCmd.Version = fmt.Sprintf("%s %s\n", version, gitCommit)

	rootCmd.AddCommand(
		ReadSecretCommand(getRootValues),
		WriteSecretCommand(getRootValues),
		TeardownCommand(getRootValues),
		ConfigCommand(getRootValues),
		APICommand(getRootValues),
		TakeoffCommand(getRootValues),
		FileCommand(getRootValues),
		StartCommand(getRootValues),
		NodeCommand(getRootValues),
	)

	if err := rootCmd.Execute(); err != nil {
		monitor.Error(mntr.ToUserError(err))
		os.Exit(1)
	}
}
