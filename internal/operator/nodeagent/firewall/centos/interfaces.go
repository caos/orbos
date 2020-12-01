package centos

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

func getEnsureInterfaces(monitor mntr.Monitor, zoneName string, current *common.ZoneDesc, desired common.Firewall) ([]string, error) {
	ensureIfaces := make([]string, 0)

	ifaces, err := getInterfaces(monitor, zoneName)
	if err != nil {
		return ensureIfaces, err
	}
	current.Interfaces = ifaces

	zone := desired.Zones[zoneName]
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
	return ensureIfaces, nil
}

func getInterfaces(monitor mntr.Monitor, zone string) ([]string, error) {
	return listFirewall(monitor, zone, "--list-interfaces")
}
