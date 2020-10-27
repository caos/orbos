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
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) (common.Current, func() error, error) {
		ensurers := make([]func() error, 0)
		current := make(common.Current, 0)

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

		current.Sort()

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

	ifaces, err := getInterfaces(monitor, name)
	if err != nil {
		return current, nil, err
	}
	current.Interfaces = ifaces

	sources, err := getSources(monitor, name)
	if err != nil {
		return current, nil, err
	}
	current.Sources = sources

	alreadyOpen, err := getPorts(monitor, name)
	if err != nil {
		return current, nil, err
	}

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

	ensureIfaces := make([]string, 0)
	zone := desired.Zones[name]
	if zone.Interfaces != nil && len(zone.Interfaces) > 0 {
		for _, iface := range zone.Interfaces {
			foundIface := false
			if current.Interfaces != nil && len(current.Interfaces) > 0 {
				for _, currentIface := range current.Interfaces {
					if currentIface == iface {
						foundIface = true
					}
				}
			}
			if !foundIface {
				ensureIfaces = append(ensureIfaces, fmt.Sprintf("--change-interface=%s", iface))
			}
		}
	}

	addSources := make([]string, 0)
	removeSources := make([]string, 0)
	if zone.Sources != nil && len(zone.Sources) > 0 {
		for _, source := range zone.Sources {
			foundSource := false
			if current.Sources != nil && len(current.Sources) > 0 {
				for _, currentSource := range current.Sources {
					if currentSource == source {
						foundSource = true
					}
				}
			}
			if !foundSource {
				addSources = append(addSources, fmt.Sprintf("--add-source=%s", source))
			}
		}
	}
	if current.Sources != nil && len(current.Sources) > 0 {
		for _, currentSource := range current.Sources {
			foundSource := false
			if zone.Sources != nil && len(zone.Sources) > 0 {
				for _, source := range zone.Sources {
					if source == currentSource {
						foundSource = true
					}
				}
			}
			if !foundSource {
				removeSources = append(removeSources, fmt.Sprintf("--remove-source=%s", currentSource))
			}
		}
	}

	if (addPorts == nil || len(addPorts) == 0) &&
		(removePorts == nil || len(removePorts) == 0) &&
		(addSources == nil || len(addSources) == 0) &&
		(removeSources == nil || len(removeSources) == 0) &&
		(ensureIfaces == nil || len(ensureIfaces) == 0) {
		return current, nil, nil
	}

	zoneName := name
	return current, func() error {
		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureIfaces, zoneName))
		if err := ensure(monitor, ensureIfaces, zoneName); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addSources, zoneName))
		if err := ensure(monitor, addSources, zoneName); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removeSources, zoneName))
		if err := ensure(monitor, removeSources, zoneName); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addPorts, zoneName))
		if err := ensure(monitor, addPorts, zoneName); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removePorts, zoneName))
		return ensure(monitor, removePorts, zoneName)
	}, nil
}

func getInterfaces(monitor mntr.Monitor, zone string) ([]string, error) {
	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("firewall-cmd", "--list-interfaces", "--zone", zone)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "running firewall-cmd --list-interfaces in order to get connected interfaces failed with stderr %s", errBuf.String())
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		fmt.Println(stdout)
	}

	return strings.Fields(stdout), nil
}

func getSources(monitor mntr.Monitor, zone string) ([]string, error) {
	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("firewall-cmd", "--list-sources", "--zone", zone)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "running firewall-cmd --list-sources in order to get the already defined firewall sources failed with stderr %s", errBuf.String())
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		fmt.Println(stdout)
	}

	return strings.Fields(stdout), nil
}

func getPorts(monitor mntr.Monitor, zone string) ([]string, error) {
	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("firewall-cmd", "--list-ports", "--zone", zone)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "running firewall-cmd --list-ports in order to get the already open firewalld ports failed with stderr %s", errBuf.String())
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		fmt.Println(stdout)
	}

	return strings.Fields(stdout), nil
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

func changeInterface(monitor mntr.Monitor, zone string, iface string) error {
	errBuf := new(bytes.Buffer)
	errBuf.Reset()
	cmd := exec.Command("firewall-cmd", "--permanent", "--zone="+zone, "--change-interface="+iface)
	cmd.Stderr = errBuf
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	if err := errors.Wrapf(cmd.Run(), "running firewall-cmd --change-interface failed with stderr %s", errBuf.String()); err != nil {
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
