package centos

import (
	"github.com/caos/orbos/internal/operator/common"
)

func getEnsureMasquerade(

	zoneName string,
	current *common.ZoneDesc,
	desired common.Firewall,
	currentZone Zone,
) string {
	ensureMasquerade := ""

	zone := desired.Zones[zoneName]
	current.Masquerade = currentZone.Masquerade

	if currentZone.Masquerade != zone.Masquerade {
		if zone.Masquerade {
			ensureMasquerade = "--add-masquerade"
		} else {
			ensureMasquerade = "--remove-masquerade"
		}
	}

	return ensureMasquerade
}
