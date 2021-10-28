package centos

import (
	"errors"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

func getEnsureMasquerade(

	monitor mntr.Monitor,
	zoneName string,
	current *common.ZoneDesc,
	desired common.Firewall,
) (
	string,
	error,
) {
	ensureMasquerade := ""

	masq, err := queryMasquerade(monitor, zoneName)
	if err != nil {
		return ensureMasquerade, err
	}

	zone := desired.Zones[zoneName]
	current.Masquerade = masq

	if masq != zone.Masquerade {
		if zone.Masquerade {
			ensureMasquerade = "--add-masquerade"
		} else {
			ensureMasquerade = "--remove-masquerade"
		}
	}

	return ensureMasquerade, nil
}

func queryMasquerade(monitor mntr.Monitor, zone string) (bool, error) {
	response, err := listFirewall(monitor, zone, "--list-all")

	if err != nil {
		return false, err
	}

	trimmed := strings.TrimSpace(strings.Join(response, " "))
	if strings.Contains(trimmed, "masquerade: yes") {
		return true, nil
	} else if strings.Contains(trimmed, "masquerade: no") {
		return false, nil
	}

	return false, errors.New("response to query not processable")
}
