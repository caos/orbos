package chore

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Run(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, os.Environ()...)
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("executing %s failed: %s", strings.Join(cmd.Args, " "), err.Error()))
	}
}
