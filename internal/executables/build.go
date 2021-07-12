//go:generate goderive .

package executables

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/caos/orbos/internal/helpers"

	"github.com/pkg/errors"
)

var Unpack func(string)

type Buildable struct {
	MainDir string
	OutDir  string
	Env     map[string]string
}

type BuiltTuple func() (bin Buildable, err error)

func Build(debug bool, gitCommit, version, githubClientID, githubClientSecret, sentryDsn string, bins ...Buildable) <-chan BuiltTuple {
	return deriveFmapBuild(curryBuild(debug, gitCommit, version, githubClientID, githubClientSecret, sentryDsn), toBuildableChan(bins))
}

func PackableBuilds(builds <-chan BuiltTuple) <-chan PackableTuple {
	return deriveFmapPackableFromBuild(packableBuild, builds)
}

func PackableFiles(paths <-chan string) <-chan PackableTuple {
	return deriveFmapPackableFromFile(packableFile, paths)
}

func packableFile(path string) PackableTuple {

	file, err := os.Open(path)
	return deriveTuplePackable(&packable{
		key:  filepath.Base(path),
		data: file,
	}, err)
}

func packableBuild(built BuiltTuple) PackableTuple {
	bin, err := built()

	file, openErr := os.Open(bin.OutDir)
	err = helpers.Concat(err, openErr)

	return deriveTuplePackable(&packable{
		key:  filepath.Base(bin.MainDir),
		data: file,
	}, err)
}

func curryBuild(debug bool, gitCommit, version, githubClientID, githubClientSecret, sentryDsn string) func(bin Buildable) BuiltTuple {
	debugCurried := deriveCurryDebug(build)(debug)
	commitCurried := deriveCurryCommit(debugCurried)(gitCommit)
	tagCurried := deriveCurryTag(commitCurried)(version)
	githubClientIDCurried := deriveCurryGithubClientID(tagCurried)(githubClientID)
	githubClientSecretCurried := deriveCurryGithubClientSecret(githubClientIDCurried)(githubClientSecret)
	sentryDsnCurried := deriveCurrySentryDSN(githubClientSecretCurried)(sentryDsn)
	return sentryDsnCurried
}

func toBuildableChan(bins []Buildable) <-chan Buildable {
	binChan := make(chan Buildable, 0)
	go func() {
		for _, bin := range bins {
			binChan <- bin
		}
		close(binChan)
	}()
	return binChan
}

func build(debug bool, gitCommit, version, githubClientID, githubClientSecret, sentryDsn string, bin Buildable) (bt BuiltTuple) {

	defer func() {
		if _, err := bt(); err != nil {
			fmt.Printf("Building %s failed\n", bin.OutDir)
			return
		}
		fmt.Printf("Successfully built %s\n", bin.OutDir)
	}()

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

	ldflags = ldflags + fmt.Sprintf("-X main.gitCommit=%s -X main.version=%s -X main.githubClientID=%s -X main.githubClientSecret=%s -X main.caosSentryDsn=%s", gitCommit, version, githubClientID, githubClientSecret, sentryDsn)

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

func builtTupleFunc(bin Buildable) func(error) BuiltTuple {
	return func(err error) BuiltTuple {
		return deriveTupleBuilt(bin, err)
	}
}

func selfPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
