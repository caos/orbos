package main

import (
	"fmt"
	"github.com/caos/orbiter/cmd/orbctl/cmds"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	rootCmd, rootValues := cmds.RootCommand()
	rootCmd.AddCommand(
		cmds.ReadSecretCommand(rootValues),
		cmds.WriteSecretCommand(rootValues),
		cmds.EditCommand(rootValues),
		cmds.TakeoffCommand(rootValues),
		cmds.TeardownCommand(rootValues),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
