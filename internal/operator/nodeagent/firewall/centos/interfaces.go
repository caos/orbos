package centos

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/common"
)

func getEnsureAndRemoveInterfaces(zoneName string, current *common.ZoneDesc, desired common.Firewall) ([]string, []string) {

	ensureIfaces := make([]string, 0)
	removeIfaces := make([]string, 0)
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
	if current.Interfaces != nil && len(current.Interfaces) > 0 {
		for _, currentIface := range current.Interfaces {
			foundIface := false
			if zone.Interfaces != nil && len(zone.Interfaces) > 0 {
				for _, iface := range zone.Interfaces {
					if iface == currentIface {
						foundIface = true
					}
				}
			}
			if !foundIface {
				removeIfaces = append(removeIfaces, fmt.Sprintf("--remove-interface=%s", currentIface))
			}
		}
	}

	return ensureIfaces, removeIfaces
}
