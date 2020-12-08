package centos

import (
	"bytes"
	"fmt"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strings"
)

const (
	prefix string = "orbos"
)

func Ensurer(monitor mntr.Monitor) nodeagent.NetworkingEnsurer {
	return nodeagent.NetworkingEnsurerFunc(func(desired common.Networking) (common.NetworkingCurrent, func() error, error) {
		current := make(common.NetworkingCurrent, 0)
		ensurers := make([]func() error, 0)

		ensurer, err := ensureInterfaces(monitor, &desired, &current)
		if err != nil {
			return current, ensurer, err
		}
		ensurers = append(ensurers, ensurer)

		return current, func() error {
			monitor.Debug("Ensuring networking")
			for _, ensurer := range ensurers {
				if err := ensurer(); err != nil {
					return err
				}
			}
			return nil
		}, nil
	})
}

func ensureInterfaces(
	monitor mntr.Monitor,
	desired *common.Networking,
	current *common.NetworkingCurrent,
) (
	func() error,
	error,
) {
	ensurers := make([]func() error, 0)
	changes := []string{}

	if desired.Interfaces == nil {
		desired.Interfaces = make(map[string]*common.NetworkingInterface, 0)
	}

	interfaces, err := queryExisting()
	if err != nil {
		return nil, err
	}

addLoop:
	for ifaceName := range desired.Interfaces {
		iface := desired.Interfaces[ifaceName]
		if iface == nil {
			return nil, errors.New("void interface")
		}
		//ensure ips for every desired interface
		ifaceNameWithPrefix := prefix + ifaceName
		ensureFunc, err := ensureInterface(monitor, ifaceNameWithPrefix, iface)
		if err != nil {
			return nil, err
		}

		if ensureFunc != nil {
			ensurers = append(ensurers, ensureFunc)
		}

		for _, alreadyIface := range interfaces {
			if alreadyIface == ifaceName {
				continue addLoop
			}
		}

		changes = append(changes, fmt.Sprintf("link add %s type %s", ifaceNameWithPrefix, iface.Type))
	}

deleteLoop:
	for _, ifaceName := range interfaces {
		if ifaceName == "" {
			continue
		}
		ifaceNameWithPrefix := prefix + ifaceName
		ipsByte, err := queryExistingInterface(ifaceNameWithPrefix)
		if err != nil {
			return nil, err
		}
		actualIps := bytes.Split(ipsByte, []byte("\n"))
		ips := make(common.MarshallableSlice, 0)
		for _, actualIp := range actualIps {
			if string(actualIp) != "" {
				ips = append(ips, string(actualIp))
			}
		}

		*current = append(*current, &common.NetworkingInterfaceCurrent{
			Name: ifaceName,
			IPs:  ips,
		})

		for desiredIfaceName := range desired.Interfaces {
			if strings.TrimPrefix(ifaceName, prefix) == desiredIfaceName {
				continue deleteLoop
			}
		}
		changes = append(changes, fmt.Sprintf("link delete %s", ifaceName))
	}

	current.Sort()
	return func() error {
		monitor.Debug(fmt.Sprintf("Ensuring part of networking"))
		if err := ensureIP(monitor, changes); err != nil {
			return err
		}

		for ifaceName := range desired.Interfaces {
			iface := desired.Interfaces[ifaceName]

			//ensure ips for every desired interface
			ifaceNameWithPrefix := prefix + ifaceName
			ensureFunc, err := ensureInterface(monitor, ifaceNameWithPrefix, iface)
			if err != nil {
				return err
			}

			if ensureFunc != nil {
				if err := ensureFunc(); err != nil {
					return err
				}
			}
		}
		return nil
	}, nil
}

func ensureInterface(
	monitor mntr.Monitor,
	name string,
	desired *common.NetworkingInterface,
) (
	func() error,
	error,
) {

	changes := []string{}

	fullInterface, err := queryExistingInterface(name)
	addedVIPs := make([][]byte, 0)
	if err == nil {
		addedVIPs = bytes.Split(fullInterface, []byte("\n"))
	} else if fullInterface != nil && len(fullInterface) == 0 {
		return nil, err
	}

addLoop:
	for idx := range desired.IPs {
		ip := desired.IPs[idx]
		if ip == "" {
			return nil, errors.New("void ip")
		}
		for idx := range addedVIPs {
			already := addedVIPs[idx]
			if string(already) == ip {
				continue addLoop
			}
		}
		if !bytes.Contains(fullInterface, []byte(ip)) {
			changes = append(changes, fmt.Sprintf("addr add %s/32 dev %s", ip, name))
		}
	}

deleteLoop:
	for idx := range addedVIPs {
		added := string(addedVIPs[idx])
		if added == "" {
			continue
		}

		for idx := range desired.IPs {
			ip := desired.IPs[idx]
			if added == ip {
				continue deleteLoop
			}
		}
		changes = append(changes, fmt.Sprintf("addr delete %s/32 dev %s", added, name))
	}

	return func() error {
		monitor.Debug(fmt.Sprintf("Ensuring part of networking with interface %s", name))
		return ensureIP(monitor, changes)
	}, nil
}

func queryExisting() ([]string, error) {
	cmd := exec.Command("/bin/sh", "-c", `ip link show | awk 'NR % 2 == 1'`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	interfaceNames := []string{}
	interfaces := strings.Split(string(output), "\n")
	for _, iface := range interfaces {
		if iface == "" {
			continue
		}

		parts := strings.Split(iface, ":")
		if len(parts) > 1 {
			name := strings.TrimSpace(parts[1])
			if strings.HasPrefix(name, prefix) {
				interfaceNames = append(interfaceNames, strings.TrimPrefix(name, prefix))
			}
		}
	}
	return interfaceNames, nil
}

func queryExistingInterface(interfaceName string) ([]byte, error) {
	cmdStr := fmt.Sprintf(`set -o pipefail && ip address show %s | grep %s | tail -n +2 | awk '{print $2}' | cut -d "/" -f 1`, interfaceName, interfaceName)

	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	return cmd.CombinedOutput()
}

func ensureIP(monitor mntr.Monitor, changes []string) (err error) {
	defer func() {
		if err == nil {
			monitor.Debug("networking changed")
		} else {
			monitor.Error(err)
		}
	}()
	cmdStr := "true"
	for _, change := range changes {
		cmdStr += fmt.Sprintf(" && sudo ip %s", change)
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	if len(changes) == 0 {
		return nil
	}

	errBuf.Reset()
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	cmd.Stderr = errBuf

	if monitor.IsVerbose() {
		fmt.Println(cmdStr)
		cmd.Stdout = os.Stdout
	}

	return errors.Wrapf(cmd.Run(), "running %s failed with stderr %s", cmdStr, errBuf.String())
}
