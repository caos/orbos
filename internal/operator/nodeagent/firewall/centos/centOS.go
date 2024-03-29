package centos

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/mntr"
)

func Ensurer(monitor mntr.Monitor, open []string) nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) (common.FirewallCurrent, func() error, error) {
		ensurers := make([]func() error, 0)
		current := make(common.FirewallCurrent, 0)

		if desired.Zones == nil {
			desired.Zones = make(map[string]*common.Zone, 0)
		}

		_, inactiveErr := runCommand(monitor, "systemctl", "is-active", "firewalld")
		_, disabledErr := runCommand(monitor, "systemctl", "is-enabled", "firewalld")
		if inactiveErr != nil || disabledErr != nil {
			monitor.WithFields(
				map[string]interface{}{
					"disabled": strconv.FormatBool(disabledErr != nil),
					"inactive": strconv.FormatBool(inactiveErr != nil),
				},
			).Info("Firewall is inactive or disabled")
			return current, func() error {
				monitor.Info("Enabling and starting firewall")
				if _, err := runCommand(monitor, "systemctl", "enable", "firewalld"); err != nil {
					return err
				}

				_, err := runCommand(monitor, "systemctl", "start", "firewalld")
				return err
			}, nil
		}

		// Ensure that all runtime config made in the previous iteration becomes permanent.
		if _, err := runFirewallCommand(monitor, "--runtime-to-permanent"); err != nil {
			return current, nil, err
		}

		currentFirewall, err := queryCurrentFirewall(monitor)
		if err != nil {
			return current, nil, err
		}

		for name, _ := range desired.Zones {
			currentZone, ensureFunc, err := ensureZone(monitor, name, desired, currentFirewall, open)
			if err != nil {
				return current, nil, err
			}
			current = append(current, currentZone)
			if ensureFunc != nil {
				ensurers = append(ensurers, ensureFunc)
			}
		}

		if len(ensurers) == 0 {
			monitor.Debug("Not changing firewall")
			return current, nil, nil
		}

		current.Sort()

		return current, func() (err error) {
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

func ensureZone(monitor mntr.Monitor, zoneName string, desired common.Firewall, currentFW map[string]Zone, open []string) (*common.ZoneDesc, func() error, error) {
	current := &common.ZoneDesc{
		Name:       zoneName,
		Interfaces: []string{},
		Services:   []*common.Service{},
		FW:         []*common.Allowed{},
	}

	current.Interfaces = currentFW[zoneName].Interfaces.slice
	current.Sources = currentFW[zoneName].Sources.slice

	ensureMasquerade := getEnsureMasquerade(zoneName, current, desired, currentFW[zoneName])
	addPorts, removePorts := getAddAndRemovePorts(current, desired.Ports(zoneName), open, currentFW[zoneName])
	ensureIfaces, removeIfaces := getEnsureAndRemoveInterfaces(zoneName, current, desired)
	addSources, removeSources := getAddAndRemoveSources(monitor, zoneName, current, desired)
	ensureTarget := getEnsureTarget(currentFW[zoneName])

	monitor.WithFields(map[string]interface{}{
		"open":  strings.Join(addPorts, ";"),
		"close": strings.Join(removePorts, ";"),
	}).Debug("firewall changes determined")

	if len(addPorts) == 0 &&
		len(removePorts) == 0 &&
		len(addSources) == 0 &&
		len(removeSources) == 0 &&
		len(ensureIfaces) == 0 &&
		len(removeIfaces) == 0 &&
		len(ensureTarget) == 0 {
		return current, nil, nil
	}

	zoneNameCopy := zoneName
	return current, func() (err error) {

		if len(ensureTarget) > 0 {

			monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureTarget, zoneNameCopy))
			if err := ensure(monitor, ensureTarget, zoneNameCopy); err != nil {
				return err
			}

			// this is the only property that needs a firewall reload
			_, err := runFirewallCommand(monitor, "--reload")
			return err
		}

		if ensureMasquerade != "" {
			monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureMasquerade, zoneNameCopy))
			if err := ensure(monitor, []string{ensureMasquerade}, zoneNameCopy); err != nil {
				return err
			}
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removeIfaces, zoneNameCopy))
		if err := ensure(monitor, removeIfaces, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureIfaces, zoneNameCopy))
		if err := ensure(monitor, ensureIfaces, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removeSources, zoneNameCopy))
		if err := ensure(monitor, removeSources, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addSources, zoneNameCopy))
		if err := ensure(monitor, addSources, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removePorts, zoneNameCopy))
		if err := ensure(monitor, removePorts, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addPorts, zoneNameCopy))
		return ensure(monitor, addPorts, zoneNameCopy)
	}, nil
}

func ensure(monitor mntr.Monitor, changes []string, zone string) error {
	if changes == nil || len(changes) == 0 {
		return nil
	}

	return changeFirewall(monitor, changes, zone)
}

func changeFirewall(monitor mntr.Monitor, changes []string, zone string) error {
	if len(changes) == 0 {
		return nil
	}

	_, err := runFirewallCommand(monitor.Verbose(), append([]string{"--zone", zone}, changes...)...)
	return err
}

func listFirewall(monitor mntr.Monitor, zone string, arg string) ([]string, error) {

	out, err := runFirewallCommand(monitor, "--zone", zone, arg)
	return strings.Fields(out), err
}

func runFirewallCommand(monitor mntr.Monitor, args ...string) (string, error) {
	return runCommand(monitor, "firewall-cmd", args...)
}

func runCommand(monitor mntr.Monitor, binary string, args ...string) (string, error) {

	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command(binary, args...)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	fullCmd := fmt.Sprintf("'%s'", strings.Join(cmd.Args, "' '"))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(`running %s failed with stderr %s: %w`, fullCmd, errBuf.String(), err)
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(fullCmd)
		fmt.Println(stdout)
	}

	return strings.TrimSuffix(stdout, "\n"), nil
}
