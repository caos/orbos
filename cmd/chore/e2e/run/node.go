package main

import (
	"context"
	"fmt"
)

func someMasterNodeContextAndID(ctx context.Context, settings programSettings, orbctl newOrbctlCommandFunc) (string, string, error) {

	var (
		context = fmt.Sprintf("%s.management", settings.orbID)
		id      string
	)
	return context, id, runCommand(settings, true, nil, func(line string) {
		id = line
	}, orbctl(ctx), "--gitops", "nodes", "list", "--context", context, "--column", "id")
}
