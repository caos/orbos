package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
)

func teardownCommand(rv rootValues) *cobra.Command {

	var (
		cmd = &cobra.Command{
			Use:   "teardown",
			Short: "Tear down an Orb",
			Long:  "Destroys a whole Orb",
			Aliases: []string{
				"shoot",
				"destroy",
				"devastate",
				"annihilate",
				"crush",
				"bulldoze",
				"total",
				"smash",
				"decimate",
				"kill",
				"trash",
				"wipe-off-the-map",
				"pulverize",
				"take-apart",
				"destruct",
				"obliterate",
				"disassemble",
				"explode",
				"blow-up",
			},
		}
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		_, logger, gitClient, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		logger.WithFields(map[string]interface{}{
			"version": version,
			"commit":  gitCommit,
			"repoURL": orbFile.URL,
		}).Info(false, "Shooting down Orbiter")

		fmt.Println("Are you absolutely sure you want to destroy all clusters and providers in this Orb? [y/N]")
		var response string
		fmt.Scanln(&response)

		if !contains([]string{"y", "yes"}, strings.ToLower(response)) {
			logger.Info(false, "Leaving Orbiter above")
			return nil
		}

		return orbiter.Destroy(
			logger,
			gitClient,
			orb.AdaptFunc(
				orbFile,
				gitCommit,
				true,
				false))
	}
	return cmd
}

func contains(slice []string, elem string) bool {
	for _, item := range slice {
		if elem == item {
			return true
		}
	}
	return false
}
