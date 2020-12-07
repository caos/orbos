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
)

func Ensurer(monitor mntr.Monitor) nodeagent.NetworkingEnsurer {
	return nodeagent.NetworkingEnsurerFunc(func(desired common.Networking) (common.NetworkingCurrent, func() error, error) {
		ensurers := make([]func() error, 0)
		current := make(common.NetworkingCurrent, 0)

		if desired.Interfaces == nil {
			desired.Interfaces = make(map[string]*common.NetworkingInterface, 0)
		}

		for name, iface := range desired.Interfaces {
			currentInterface, ensureFunc, err := ensureInterface(monitor, name, iface)
			if err != nil {
				return current, nil, err
			}
			current = append(current, currentInterface)
			if ensureFunc != nil {
				ensurers = append(ensurers, ensureFunc)
			}
		}

		current.Sort()
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

func ensureInterface(
	monitor mntr.Monitor,
	name string,
	desired *common.NetworkingInterface,
) (
	*common.NetworkingInterfaceCurrent,
	func() error,
	error,
) {

	current := &common.NetworkingInterfaceCurrent{
		Name: name,
		IPs:  nil,
	}
	changes := []string{}

	fullInterface, err := queryExistingInterface(name)
	addedVIPs := bytes.Split(fullInterface, []byte("\n"))
	if err != nil {
		if addedVIPs != nil && len(addedVIPs) == 0 {
			changes = append(changes, fmt.Sprintf("link add %s type dummy", name))
		} else {
			return nil, nil, err
		}
	}

addLoop:
	for idx := range desired.IPs {
		ip := desired.IPs[idx]
		if ip == "" {
			return nil, nil, errors.New("void ip")
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

		current.IPs = append(current.IPs, added)

		for idx := range desired.IPs {
			ip := desired.IPs[idx]
			if added == ip {
				continue deleteLoop
			}
		}
		changes = append(changes, fmt.Sprintf("addr delete %s/32 dev %s", added, name))
	}

	return current, func() error {
		monitor.Debug(fmt.Sprintf("Ensuring part of networking with interface %s", name))
		return ensureIP(monitor, changes)
	}, nil
}

func queryExistingInterface(interfaceName string) ([]byte, error) {
	cmdStr := fmt.Sprintf(`INNEROUT="$(set -o pipefail && sudo ip address show %s | grep %s | tail -n +2 | awk '{print $2}' | cut -d "/" -f 1)" && echo $INNEROUT`, interfaceName, interfaceName)

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	errBuf.Reset()

	cmd := exec.Command(cmdStr)
	cmd.Stderr = errBuf

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
	cmd := exec.Command(cmdStr)
	cmd.Stderr = errBuf

	if monitor.IsVerbose() {
		fmt.Println(cmdStr)
		cmd.Stdout = os.Stdout
	}

	return errors.Wrapf(cmd.Run(), "running %s failed with stderr %s", cmdStr, errBuf.String())
}
