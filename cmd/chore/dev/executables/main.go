package main

import (
	"github.com/caos/orbos/v5/cmd/chore"
)

func main() {
	if err := chore.BuildExecutables(true, true); err != nil {
		panic(err)
	}
}
