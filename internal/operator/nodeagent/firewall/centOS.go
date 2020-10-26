package firewall

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/mntr"
)

func centosEnsurer(monitor mntr.Monitor, ignore []string) nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) ([]*common.ZoneDesc, func() error, error) {
		ensurers := make([]func() error, 0)
		current := make([]*common.ZoneDesc, 0)

		if desired.Zones == nil {
			desired.Zones = make(map[string]*common.Zone, 0)
		}

		for name, _ := range desired.Zones {
			currentZone, ensureFunc, err := ensureZone(monitor, name, desired, ignore)
			if err != nil {
				return current, nil, err
			}
			current = append(current, currentZone)
			if ensureFunc != nil {
				ensurers = append(ensurers, ensureFunc)
			}
		}

		cmd := exec.Command("systemctl", "is-active", "firewalld")
		if monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}

		if cmd.Run() != nil || len(ensurers) == 0 {
			monitor.Debug("Not changing firewall")
			return current, nil, nil
		}

		return current, func() error {
			monitor.Debug("Ensuring firewall")
			for _, ensurer := range ensurers {
				if err := ensurer(); err != nil {
					return err
				}
			}
			return nil
		}, nil
	})
}

func ensureZone(monitor mntr.Monitor, name string, desired common.Firewall, ignore []string) (*common.ZoneDesc, func() error, error) {
	current := &common.ZoneDesc{
		Name:       name,
		Interfaces: []string{},
		Services:   []*common.Service{},
		FW:         []*common.Allowed{},
	}

	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("firewall-cmd", "--list-ports", "--zone", name)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	if err := cmd.Run(); err != nil {
		return current, nil, errors.Wrapf(err, "running firewall-cmd --list-ports in order to get the already open firewalld ports failed with stderr %s", errBuf.String())
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		fmt.Println(stdout)
	}

	alreadyOpen := strings.Fields(stdout)
	addPorts := make([]string, 0)
	removePorts := make([]string, 0)

	ensureOpen := append(desired.Ports(name), ignoredPorts(ignore)...)
openloop:
	for _, des := range ensureOpen {
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
		fields := strings.Split(already, "/")
		port := fields[0]
		protocol := fields[1]

		current.FW = append(current.FW, &common.Allowed{
			Port:     port,
			Protocol: protocol,
		})

		for _, des := range ensureOpen {
			if des.Port == port && des.Protocol == protocol {
				continue closeloop
			}
		}
		removePorts = append(removePorts, fmt.Sprintf("--remove-port=%s", already))
	}

	monitor.WithFields(map[string]interface{}{
		"open":  strings.Join(addPorts, ";"),
		"close": strings.Join(removePorts, ";"),
	}).Debug("firewall changes determined")

	if (addPorts == nil || len(addPorts) == 0) && (removePorts == nil || len(removePorts) == 0) {
		return current, nil, nil
	}

	zoneName := name
	return current, func() error {
		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addPorts, zoneName))
		if err := ensure(monitor, addPorts, zoneName); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removePorts, zoneName))
		return ensure(monitor, removePorts, zoneName)
	}, nil
}

func ignoredPorts(ports []string) []*common.Allowed {
	allowed := make([]*common.Allowed, len(ports))
	for idx, port := range ports {
		allowed[idx] = &common.Allowed{
			Port:     port,
			Protocol: "tcp",
		}
	}
	return allowed
}

func ensure(monitor mntr.Monitor, changes []string, zone string) error {
	if changes == nil || len(changes) == 0 {
		return nil
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("systemctl", "enable", "firewalld")
	cmd.Stderr = errBuf

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
	cmd.Stderr = errBuf

	fullCmd = strings.Join(cmd.Args, " ")
	if monitor.IsVerbose() {
		fmt.Println(fullCmd)
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	return changeFirewall(monitor, changes, zone)
}

func changeFirewall(monitor mntr.Monitor, changes []string, zone string) (err error) {

	changesMonitor := monitor.WithField("changes", strings.Join(changes, ";"))
	changesMonitor.Debug("Changing firewall")

	defer func() {
		if err == nil {
			changesMonitor.Debug("firewall changed")
		} else {
			changesMonitor.Error(err)
		}
	}()

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	if len(changes) == 0 {
		return nil
	}

	errBuf.Reset()
	cmd := exec.Command("firewall-cmd", append([]string{"--permanent", "--zone", zone}, changes...)...)
	cmd.Stderr = errBuf

	fullCmd := strings.Join(cmd.Args, " ")
	if monitor.IsVerbose() {
		fmt.Println(fullCmd)
		cmd.Stdout = os.Stdout
	}

	if err := errors.Wrapf(cmd.Run(), "running %s failed with stderr %s", fullCmd, errBuf.String()); err != nil {
		return err
	}

	return reloadFirewall(monitor)
}

func reloadFirewall(monitor mntr.Monitor) error {
	errBuf := new(bytes.Buffer)
	errBuf.Reset()
	cmd := exec.Command("firewall-cmd", "--reload")
	cmd.Stderr = errBuf
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	return errors.Wrapf(cmd.Run(), "running firewall-cmd --reload failed with stderr %s", errBuf.String())
}
