//go:generate goderive .

package executables

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

var Unpack func(string)

type BuiltTuple func() (mainDir string, executable io.Reader, close func(), err error)

func Build(debug bool, gitCommit string, gitTag string, mainDir ...string) <-chan BuiltTuple {
	return deriveFmap(curryBuild(debug, gitCommit, gitTag), toChan(mainDir))
}

func curryBuild(debug bool, gitCommit string, gitTag string) func(mainDir string) BuiltTuple {
	debugCurried := deriveCurryDebug(build)(debug)
	commitCurried := deriveCurryCommit(debugCurried)(gitCommit)
	tagCurried := deriveCurryTag(commitCurried)(gitTag)
	return tagCurried
}

func toChan(strs []string) <-chan string {
	strChan := make(chan string, 0)
	go func() {
		for _, str := range strs {
			strChan <- str
		}
		close(strChan)
	}()
	return strChan
}

func build(debug bool, gitCommit, gitTag, mainDir string) BuiltTuple {

	base := filepath.Base(mainDir)
	ext := filepath.Ext(base)
	bf := filepath.Join(os.TempDir(), base[0:len(base)-len(ext)]) + ".build"

	args := []string{"build", "-o", bf}

	ldflags := "-s -w "
	if debug {
		ldflags = ""
		args = append(args,
			"-gcflags",
			"all=-N -l")
	}

	ldflags = ldflags + fmt.Sprintf("-X main.gitCommit=%s -X main.gitTag=%s", gitCommit, gitTag)

	cmd := exec.Command("go", append(args, "-ldflags", ldflags, mainDir)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	builtTuple := builtTupleFunc(mainDir)

	if err := cmd.Run(); err != nil {
		return builtTuple(nil, func() {}, errors.Wrapf(cmd.Run(), "building %s failed", bf))
	}

	openFile, err := os.Open(bf)
	return builtTuple(openFile, func(f *os.File) func() {
		return func() {
			f.Close()
			os.Remove(f.Name())
		}
	}(openFile), errors.Wrapf(err, "opening %s failed", bf))
}

func builtTupleFunc(mainDir string) func(io.Reader, func(), error) BuiltTuple {
	return func(executable io.Reader, close func(), err error) BuiltTuple {
		return deriveTupleBuilt(mainDir, executable, close, err)
	}
}

func selfPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
