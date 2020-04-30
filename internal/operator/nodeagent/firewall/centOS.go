package firewall

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/mntr"
)

func centosEnsurer(monitor mntr.Monitor, ignore []string) nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) ([]*common.Allowed, func() error, error) {

		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer
		)

		cmd := exec.Command("firewall-cmd", "--list-ports")
		cmd.Stderr = &errBuf
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			return nil, nil, errors.Wrapf(err, "running firewall-cmd --list-ports in order to get the already open firewalld ports failed with stderr %s", errBuf.String())
		}

		stdout := outBuf.String()
		if monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			fmt.Println(stdout)
		}

		alreadyOpen := strings.Fields(stdout)
		addPorts := make([]string, 0)
		removePorts := make([]string, 0)
	openloop:
		for _, des := range desired {
			desStr := fmt.Sprintf("%s/%s", des.Port, des.Protocol)
			for _, already := range append(alreadyOpen, ignore...) {
				if desStr == already {
					continue openloop
				}
			}
			addPorts = append(addPorts, fmt.Sprintf("--add-port=%s", desStr))
		}

		for _, ign := range ignore {
			desired[ign] = &common.Allowed{
				Port:     ign,
				Protocol: "tcp",
			}
		}

		current := make([]*common.Allowed, len(alreadyOpen))
	closeloop:
		for idx, already := range alreadyOpen {
			fields := strings.Split(already, "/")
			port := fields[0]
			protocol := fields[1]
			current[idx] = &common.Allowed{Port: port, Protocol: protocol}
			for _, des := range desired {
				if des.Port == port && des.Protocol == protocol {
					continue closeloop
				}
			}
			removePorts = append(removePorts, fmt.Sprintf("--remove-port=%s", already))
		}

		cmd = exec.Command("systemctl", "is-active", "firewalld")
		if monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}

		monitor.WithFields(map[string]interface{}{
			"open":  strings.Join(addPorts, ";"),
			"close": strings.Join(removePorts, ";"),
		}).Debug("Firewall changes determined")

		if cmd.Run() != nil || len(addPorts) == 0 && len(removePorts) == 0 {
			monitor.Debug("Not changing firewall")
			return current, nil, nil
		}

		return current, func() error {

			errBuf.Reset()
			cmd = exec.Command("systemctl", "enable", "firewalld")
			cmd.Stderr = &errBuf

			fullCmd := strings.Join(cmd.Args, " ")
			if monitor.IsVerbose() {
				fmt.Println(fullCmd)
				cmd.Stdout = os.Stdout
			}

			if err := cmd.Run(); err != nil {
				return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
			}

			errBuf.Reset()
			cmd = exec.Command("systemctl", "start", "firewalld")
			cmd.Stderr = &errBuf

			fullCmd = strings.Join(cmd.Args, " ")
			if monitor.IsVerbose() {
				fmt.Println(fullCmd)
				cmd.Stdout = os.Stdout
			}

			if err := cmd.Run(); err != nil {
				return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
			}

			if err := changeFirewall(monitor, addPorts); err != nil {
				return err
			}

			return changeFirewall(monitor, removePorts)
		}, nil
	})
}

func changeFirewall(monitor mntr.Monitor, changes []string) (err error) {

	changesMonitor := monitor.WithField("changes", strings.Join(changes, ";"))
	changesMonitor.Debug("Changing firewall")

	defer func() {
		if err == nil {
			changesMonitor.Debug("Firewall changed")
		} else {
			changesMonitor.Error(err)
		}
	}()

	var errBuf bytes.Buffer
	if len(changes) == 0 {
		return nil
	}

	errBuf.Reset()
	cmd := exec.Command("firewall-cmd", append([]string{"--permanent"}, changes...)...)
	cmd.Stderr = &errBuf

	fullCmd := strings.Join(cmd.Args, " ")
	if monitor.IsVerbose() {
		fmt.Println(fullCmd)
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	errBuf.Reset()
	cmd = exec.Command("firewall-cmd", "--reload")
	cmd.Stderr = &errBuf
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	return errors.Wrapf(cmd.Run(), "running firewall-cmd --reload failed with stderr %s", errBuf.String())
}
