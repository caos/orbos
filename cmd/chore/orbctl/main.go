package main

import (
	"os"

	"github.com/caos/orbos/cmd/chore"
)

func main() {

	var debug, skipRequild bool
	for idx, arg := range os.Args {
		switch arg {
		case "--debug":
			debug = true
			removeFromCommandArguments(idx)
		case "--skip-rebuild":
			skipRequild = true
			removeFromCommandArguments(idx)
		}
	}

	newOrbctl, err := chore.Orbctl(debug, skipRequild)
	if err != nil {
		panic(err)
	}

	cmd := newOrbctl()
	cmd.Args = append(cmd.Args, os.Args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeFromCommandArguments(idx int) {
	os.Args = append(os.Args[0:idx], os.Args[idx+1:]...)
}
