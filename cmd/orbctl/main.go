package main

import (
	"fmt"
	"os"
)

var (
	// Build arguments
	gitCommit = "none"
	version   = "none"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	rootCmd, rootValues := RootCommand()
	rootCmd.Version = fmt.Sprintf("%s %s\n", version, gitCommit)
	rootCmd.AddCommand(
		ReadSecretCommand(rootValues),
		WriteSecretCommand(rootValues),
		EditCommand(rootValues),
		TakeoffCommand(rootValues),
		TeardownCommand(rootValues),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
