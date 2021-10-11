package centos

import (
	"context"

	"github.com/caos/orbos/mntr"
)

func getEnsureTarget(ctx context.Context, monitor mntr.Monitor, zoneName string) ([]string, error) {
	var changeTarget []string
	target, err := runFirewallCommand(ctx, monitor, "--zone", zoneName, "--permanent", "--get-target")
	if err != nil {
		return changeTarget, err
	}

	if target != "default" {
		changeTarget = []string{"--permanent", "--set-target=default"}
	}
	return changeTarget, nil
}
