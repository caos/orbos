package centos

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
	"strings"
)

func getAddAndRemovePorts(
	monitor mntr.Monitor,
	zone string,
	current *common.ZoneDesc,
	desiredPorts []*common.Allowed,
	ignore []string,
) (
	[]string,
	[]string,
	error,
) {

	alreadyOpen, err := getPorts(monitor, zone)
	if err != nil {
		return nil, nil, err
	}

	addPorts := make([]string, 0)
	removePorts := make([]string, 0)

	ensureOpen := append(desiredPorts, ignoredPorts(ignore)...)
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

	return addPorts, removePorts, nil
}

func getPorts(monitor mntr.Monitor, zone string) ([]string, error) {
	return listFirewall(monitor, zone, "--list-ports")
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
