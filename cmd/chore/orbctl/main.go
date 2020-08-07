package main

import (
	"os"

	"github.com/caos/orbos/cmd/chore"
)

func main() {

	var debug bool
	for idx, arg := range os.Args {
		if arg == "--debug" {
			debug = true
			os.Args = append(os.Args[0:idx], os.Args[idx+1:]...)
			break
		}
	}

	newOrbctl, err := chore.Orbctl(debug)
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
