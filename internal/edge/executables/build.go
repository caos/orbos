//go:generate goderive .

package executables

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

var Unpack func(string)

type Bin struct {
	MainDir string
	Env     map[string]string
}

type BuiltTuple func() (bin Bin, executable io.Reader, close func(), err error)

func Build(debug bool, gitCommit, gitTag string, bins ...Bin) <-chan BuiltTuple {
	return deriveFmap(curryBuild(debug, gitCommit, gitTag), toChan(bins))
}

func curryBuild(debug bool, gitCommit, gitTag string) func(bin Bin) BuiltTuple {
	debugCurried := deriveCurryDebug(build)(debug)
	commitCurried := deriveCurryCommit(debugCurried)(gitCommit)
	tagCurried := deriveCurryTag(commitCurried)(gitTag)
	return tagCurried
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

func build(debug bool, gitCommit, gitTag string, bin Bin) BuiltTuple {

	builtTuple := builtTupleFunc(bin)

	bf, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return builtTuple(nil, func() {}, errors.Wrap(err, "opening tempfile failed"))
	}

	args := []string{"build", "-o", bf.Name()}

	ldflags := "-s -w "
	if debug {
		ldflags = ""
		args = append(args,
			"-gcflags",
			"all=-N -l")
	}

	ldflags = ldflags + fmt.Sprintf("-X main.gitCommit=%s -X main.gitTag=%s", gitCommit, gitTag)

	cmdEnv := os.Environ()
	for k, v := range bin.Env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("go", append(args, "-ldflags", ldflags, bin.MainDir)...)
	cmd.Env = cmdEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return builtTuple(bf, func() {
		bf.Close()
		os.Remove(bf.Name())
	}, errors.Wrapf(cmd.Run(), "building %s failed", bf.Name()))
}

func builtTupleFunc(bin Bin) func(io.Reader, func(), error) BuiltTuple {
	return func(executable io.Reader, close func(), err error) BuiltTuple {
		return deriveTupleBuilt(bin, executable, close, err)
	}
}

func selfPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
