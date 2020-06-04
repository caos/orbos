package main

import (
	"fmt"
	"github.com/caos/orbos/internal/stores/github"
	"math/rand"
	"os"
)

var (
	// Build arguments
	gitCommit          = "none"
	version            = "none"
	githubClientID     = "none"
	githubClientSecret = "none"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	github.ClientID = githubClientID
	github.ClientSecret = githubClientSecret
	github.Key = RandStringBytes(32)

	rootCmd, rootValues := RootCommand()
	rootCmd.Version = fmt.Sprintf("%s %s\n", version, gitCommit)

	takeoff := TakeoffCommand(rootValues)
	takeoff.AddCommand(
		StartBoom(rootValues),
		StartOrbiter(rootValues),
	)
	rootCmd.AddCommand(
		ReadSecretCommand(rootValues),
		WriteSecretCommand(rootValues),
		EditCommand(rootValues),
		TeardownCommand(rootValues),
		ConfigCommand(rootValues),
		takeoff,
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
