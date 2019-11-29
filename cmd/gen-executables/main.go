//go:generate goderive .

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/caos/orbiter/internal/edge/executables"
)

func main() {

	tag := flag.String("tag", "none", "Path to the git repositorys path to the file containing orbiters current state")
	commit := flag.String("commit", "none", "Path to the git repositorys path to the file containing orbiters current state")
	debug := flag.Bool("debug", false, "Compile executables with debugging features enabled")

	flag.Parse()

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	_, selfPath, _, _ := runtime.Caller(0)
	cmdPath := filepath.Join(filepath.Dir(selfPath), "..")

	path := curryJoinPath(cmdPath)

	if err := executables.PreBuild(executables.Build(
		*debug, *commit, *tag,
		path("nodeagent"),
		path("health"),
	)); err != nil {
		panic(err)
	}
}

func joinPath(cmdPath string, dir string) string {
	return filepath.Join(cmdPath, dir)
}

func curryJoinPath(cmdPath string) func(dir string) string {
	return func(dir string) string {
		return joinPath(cmdPath, dir)
	}
}
