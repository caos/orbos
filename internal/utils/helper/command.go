package helper

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/mntr"
)

func RunWithOutput(monitor mntr.Monitor, cmd exec.Cmd) ([]byte, error) {

	var command string
	for _, arg := range cmd.Args {
		if strings.Contains(arg, " ") {
			command += " \\\"" + arg + "\\\""
			continue
		}
		command += " " + arg
	}
	command = command[1:]

	cmdMonitor := monitor.WithFields(map[string]interface{}{
		"cmd": command,
	})

	cmdMonitor.Debug("Executing")

	out, err := cmd.CombinedOutput()
	cmdMonitor.Debug(string(out))

	if err != nil {
		return nil, fmt.Errorf("error while executing command: \"%s\": response: %s: %w", strings.Join(cmd.Args, "\" \""), string(out), err)
	}

	return out, nil
}

func Run(monitor mntr.Monitor, cmd exec.Cmd) error {
	_, err := RunWithOutput(monitor, cmd)
	return err
}
