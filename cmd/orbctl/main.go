package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/caos/orbos/internal/stores/github"
)

var (
	// Build arguments
	gitCommit          = "none"
	version            = "none"
	githubClientID     = "none"
	githubClientSecret = "none"
)

func main() {
	github.ClientID = githubClientID
	github.ClientSecret = githubClientSecret
	github.Key = RandStringBytes(32)

	rootCmd, rootValues := RootCommand()
	rootCmd.Version = fmt.Sprintf("%s %s\n", version, gitCommit)

	takeoff := TakeoffCommand(rootValues)
	takeoff.AddCommand(
		StartBoom(rootValues),
		StartOrbiter(rootValues),
		StartZitadel(rootValues),
	)

	nodes := NodeCommand()
	nodes.AddCommand(
		ReplaceCommand(rootValues),
		RebootCommand(rootValues),
		ExecCommand(rootValues),
		ListCommand(rootValues),
	)

	rootCmd.AddCommand(
		ReadSecretCommand(rootValues),
		WriteSecretCommand(rootValues),
		EditCommand(rootValues),
		TeardownCommand(rootValues),
		ConfigCommand(rootValues),
		APICommand(rootValues),
		BackupListCommand(rootValues),
		RestoreCommand(rootValues),
		BackupCommand(rootValues),
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
