//go:generate goderive .

package executables

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

var Unpack func(string)

type Bin struct {
	MainDir string
	OutDir  string
	Env     map[string]string
}

type BuiltTuple func() (bin Bin, err error)

func Build(debug bool, gitCommit, version, githubClientID, githubClientSecret string, bins ...Bin) <-chan BuiltTuple {
	return deriveFmap(curryBuild(debug, gitCommit, version, githubClientID, githubClientSecret), toChan(bins))
}

func curryBuild(debug bool, gitCommit, version, githubClientID, githubClientSecret string) func(bin Bin) BuiltTuple {
	debugCurried := deriveCurryDebug(build)(debug)
	commitCurried := deriveCurryCommit(debugCurried)(gitCommit)
	tagCurried := deriveCurryTag(commitCurried)(version)
	githubClientIDCurried := deriveCurryGithubClientID(tagCurried)(githubClientID)
	githubClientSecretCurried := deriveCurryGithubClientSecret(githubClientIDCurried)(githubClientSecret)
	return githubClientSecretCurried
}

func toChan(bins []Bin) <-chan Bin {
	binChan := make(chan Bin, 0)
	go func() {
		for _, bin := range bins {
			binChan <- bin
		}
		close(binChan)
	}()
	return binChan
}

func build(debug bool, gitCommit, version, githubClientID, githubClientSecret string, bin Bin) BuiltTuple {

	if bin.OutDir == "" {
		bin.OutDir = filepath.Join(os.TempDir(), filepath.Base(bin.MainDir))
	}

	builtTuple := builtTupleFunc(bin)

	args := []string{"build", "-o", bin.OutDir}

	ldflags := "-s -w "
	if debug {
		ldflags = ""
		args = append(args,
			"-gcflags",
			"all=-N -l")
	}

	ldflags = ldflags + fmt.Sprintf("-X main.gitCommit=%s -X main.version=%s -X main.githubclientid=%s -X main.githubclientsecret=%s", gitCommit, version, githubClientID, githubClientSecret)

	cmdEnv := os.Environ()
	for k, v := range bin.Env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("go", append(args, "-ldflags", ldflags, bin.MainDir)...)
	cmd.Env = cmdEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return builtTuple(errors.Wrapf(cmd.Run(), "building %s failed", bin.OutDir))
}

func builtTupleFunc(bin Bin) func(error) BuiltTuple {
	return func(err error) BuiltTuple {
		return deriveTupleBuilt(bin, err)
	}
}

func selfPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
