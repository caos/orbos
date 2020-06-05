package helper

import (
	"os/exec"
	"strings"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func Run(monitor mntr.Monitor, cmd exec.Cmd) error {

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

	return errors.Wrapf(err, "Error while executing command: Response: %s", string(out))
}

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

	return out, errors.Wrapf(err, "Error while executing command: Response: %s", string(out))
}
