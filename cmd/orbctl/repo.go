package main

import (
	"github.com/caos/orbos/v5/pkg/git"
	"github.com/caos/orbos/v5/pkg/orb"
)

func initRepo(orbConfig *orb.Orb, gitClient *git.Client) error {
	if err := orbConfig.IsConnectable(); err != nil {
		return err
	}

	if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
		return err
	}

	return gitClient.Clone()
}
