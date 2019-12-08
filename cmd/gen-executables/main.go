//go:generate goderive .

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/caos/orbiter/internal/edge/executables"
)

func main() {

	tag := flag.String("tag", "none", "Path to the git repositorys path to the file containing orbiters current state")
	commit := flag.String("commit", "none", "Path to the git repositorys path to the file containing orbiters current state")
	orbctldir := flag.String("orbctl", "", "Build orbctl binaries to this directory")
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
		executables.Bin{MainDir: path("nodeagent")},
		executables.Bin{MainDir: path("health")},
	)); err != nil {
		panic(err)
	}

	if *orbctldir == "" {
		return
	}

	orbctls := executables.Build(*debug, *commit, *tag,
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "darwin", "GOARCH": "386"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "darwin", "GOARCH": "amd64"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "freebsd", "GOARCH": "386"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "freebsd", "GOARCH": "amd64"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "linux", "GOARCH": "386"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "linux", "GOARCH": "amd64"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "openbsd", "GOARCH": "386"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "openbsd", "GOARCH": "amd64"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "windows", "GOARCH": "386"}},
		executables.Bin{MainDir: path("orbctl"), Env: map[string]string{"GOOS": "windows", "GOARCH": "amd64"}},
	)

	var wg sync.WaitGroup
	for orbctl := range orbctls {
		wg.Add(1)
		go writeOrbctl(&wg, *orbctldir, orbctl)
	}

	wg.Wait()
}

func curryJoinPath(cmdPath string) func(dir string) string {
	return func(dir string) string {
		return filepath.Join(cmdPath, dir)
	}
}

func writeOrbctl(wg *sync.WaitGroup, outDir string, orbctlTuple executables.BuiltTuple) {
	bin, executable, close, err := orbctlTuple()
	defer close()
	defer wg.Done()
	if err != nil {
		panic(err)
	}

	filePath := filepath.Join(outDir, fmt.Sprintf("orbctl-%s-%s", bin.Env["GOOS"], bin.Env["GOARCH"]))
	if bin.Env["GOOS"] == "windows" {
		filePath = filePath + ".exe"
	}

	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, executable)
	if err != nil {
		panic(err)
	}
}
