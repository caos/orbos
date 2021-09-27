package main

import (
	"fmt"
	"os"

	"github.com/caos/orbos/v5/mntr"
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

	start := StartCommand()
	start.AddCommand(
		StartBoom(getRootValues),
		StartOrbiter(getRootValues),
		StartNetworking(getRootValues),
	)

	file := FileCommand()
	file.AddCommand(
		EditCommand(getRootValues),
		PrintCommand(getRootValues),
		//		PatchCommand(getRootValues),
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
		TeardownCommand(getRootValues),
		ConfigCommand(getRootValues),
		APICommand(getRootValues),
		TakeoffCommand(getRootValues),
		file,
		start,
		nodes,
	)

	if err := rootCmd.Execute(); err != nil {
		monitor.Error(mntr.ToUserError(err))
		os.Exit(1)
	}
}
