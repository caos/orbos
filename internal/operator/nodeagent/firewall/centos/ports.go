package centos

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
)

func getAddAndRemovePorts(
	current *common.ZoneDesc,
	desiredPorts []*common.Allowed,
	open []string,
	currentZone Zone,
) (
	[]string,
	[]string,
) {

	ensure := make([]string, 0)
	remove := make([]string, 0)

	alwaysOpen := ignoredPorts(open)

	//ports that should stay open
	for _, open := range alwaysOpen {
		found := false
		openStr := fmt.Sprintf("%s/%s", open.Port, open.Protocol)
		for _, open := range currentZone.Ports.slice {
			if open == openStr {
				found = true
				break
			}
		}
		if !found {
			ensure = append(ensure, fmt.Sprintf("--add-port=%s", openStr))
		}
	}

	//desired ports
	for _, desired := range desiredPorts {
		found := false
		desStr := fmt.Sprintf("%s/%s", desired.Port, desired.Protocol)
		for _, open := range currentZone.Ports.slice {
			if open == desStr {
				found = true
				break
			}
		}
		if !found {
			ensure = append(ensure, fmt.Sprintf("--add-port=%s", desStr))
		}
	}

	//port that are not desired anymore
	for _, open := range currentZone.Ports.slice {
		found := false

		fields := strings.Split(open, "/")
		port := fields[0]
		protocol := fields[1]

		current.FW = append(current.FW, &common.Allowed{
			Port:     port,
			Protocol: protocol,
		})

		for _, desired := range desiredPorts {
			if desired.Port == port && desired.Protocol == protocol {
				found = true
				break
			}
		}

		if !found {
			for _, open := range alwaysOpen {
				if open.Port == port && open.Protocol == protocol {
					found = true
					break
				}
			}
		}

		if !found {
			remove = append(remove, fmt.Sprintf("--remove-port=%s", open))
		}
	}

	return ensure, remove
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
