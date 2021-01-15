package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/stores/github"
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

	defer monitor.RecoverPanic()

	github.ClientID = githubClientID
	github.ClientSecret = githubClientSecret
	github.Key = RandStringBytes(32)

	rootCmd, getRootValues := RootCommand()
	rootCmd.Version = fmt.Sprintf("%s %s\n", version, gitCommit)

	takeoff := TakeoffCommand(getRootValues)
	takeoff.AddCommand(
		StartBoom(getRootValues),
		StartOrbiter(getRootValues),
		StartDatabase(getRootValues),
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
		BackupListCommand(getRootValues),
		RestoreCommand(getRootValues),
		BackupCommand(getRootValues),
		takeoff,
		nodes,
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
