package main

import (
	"fmt"

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

	takeoff := TakeoffCommand(getRootValues)
	takeoff.AddCommand(
		StartBoom(getRootValues),
		StartOrbiter(getRootValues),
		StartNetworking(getRootValues),
	)

	nodes := NodeCommand()
	nodes.AddCommand(
		ReplaceCommand(getRootValues),
		RebootCommand(getRootValues),
		ExecCommand(getRootValues),
		ListCommand(getRootValues),
	)

	rootCmd.AddCommand(
		ReadSecretCommand(getRootValues),
		WriteSecretCommand(getRootValues),
		EditCommand(getRootValues),
		TeardownCommand(getRootValues),
		ConfigCommand(getRootValues),
		APICommand(getRootValues),
		takeoff,
		nodes,
	)

	if err := rootCmd.Execute(); err != nil {
		monitor.Error(err)
	}
}
