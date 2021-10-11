package centos

import (
	"github.com/caos/orbos/mntr"
)

func getEnsureTarget(monitor mntr.Monitor, zoneName string) ([]string, error) {
	var changeTarget []string
	target, err := runFirewallCommand(monitor, "--zone", zoneName, "--permanent", "--get-target")
	if err != nil {
		return changeTarget, err
	}

	if target != "default" {
		changeTarget = []string{"--permanent", "--set-target=default"}
	}
	return changeTarget, nil
}
