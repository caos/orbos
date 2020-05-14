package dep

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbiter/mntr"

	"github.com/pkg/errors"
)

func CurrentSysctlConfig(monitor mntr.Monitor, property string) (bool, error) {

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	cmd := exec.Command("sysctl", property)
	cmd.Stderr = &errBuf
	cmd.Stdout = &outBuf

	fullCmd := strings.Join(cmd.Args, " ")
	monitor.WithFields(map[string]interface{}{"cmd": fullCmd}).Debug("Executing")

	if err := cmd.Run(); err != nil {
		return false, errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	return outBuf.String() == fmt.Sprintf("%s = 1\n", property), nil
}
