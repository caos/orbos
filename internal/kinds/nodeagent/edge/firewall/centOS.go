package firewall

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/nodeagent/adapter"
)

func centosEnsurer(logger logging.Logger) adapter.FirewallEnsurer {
	return adapter.FirewallEnsurerFunc(func(desired operator.Firewall) error {

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		cmd := exec.Command("firewall-cmd", "--list-ports")
		cmd.Stderr = &errBuf
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "running firewall-cmd --list-ports in order to get the already open firewalld ports failed with stderr %s", errBuf.String())
		}

		stdout := outBuf.String()
		if logger.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			fmt.Println(stdout)
		}

		alreadyOpen := strings.Fields(stdout)
		addPorts := make([]string, 0)
		removePorts := make([]string, 0)
	openloop:
		for _, des := range desired {
			desStr := fmt.Sprintf("%s/%s", des.Port, des.Protocol)
			for _, already := range alreadyOpen {
				if desStr == already {
					continue openloop
				}
			}
			addPorts = append(addPorts, fmt.Sprintf("--add-port=%s", desStr))
		}

	closeloop:
		for _, already := range alreadyOpen {
			for _, des := range desired {
				if fmt.Sprintf("%s/%s", des.Port, des.Protocol) == already {
					continue closeloop
				}
			}
			removePorts = append(removePorts, fmt.Sprintf("--remove-port=%s", already))
		}

		if err := changeFirewall(logger, addPorts); err != nil {
			return err
		}

		return changeFirewall(logger, removePorts)
	})
}

func changeFirewall(logger logging.Logger, changes []string) error {
	var errBuf bytes.Buffer
	if len(changes) == 0 {
		return nil
	}

	errBuf.Reset()
	cmd := exec.Command("firewall-cmd", append([]string{"--permanent"}, changes...)...)
	cmd.Stderr = &errBuf

	fullCmd := strings.Join(cmd.Args, " ")
	if logger.IsVerbose() {
		fmt.Println(fullCmd)
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	errBuf.Reset()
	cmd = exec.Command("firewall-cmd", "--reload")
	cmd.Stderr = &errBuf
	if logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	return errors.Wrapf(cmd.Run(), "running firewall-cmd --reload failed with stderr %s", errBuf.String())
}
