package api

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/tree"
)

type PushDesiredFunc func(monitor mntr.Monitor) error

func PushOrbiterDesiredFunc(gitClient *git.Client, desired *tree.Tree) func(mntr.Monitor) error {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing orbiter desired state")
		return PushGitDesiredStates(monitor, "Orbiter desired state written", gitClient, []GitDesiredState{{
			Desired: desired,
			Path:    git.OrbiterFile,
		}})
	}
}

func PushBoomDesiredFunc(gitClient *git.Client, desired *tree.Tree) func(mntr.Monitor) error {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing boom desired state")
		return PushGitDesiredStates(monitor, "Boom desired state written", gitClient, []GitDesiredState{{
			Desired: desired,
			Path:    git.BoomFile,
		}})
	}
}

func PushNetworkingDesiredFunc(gitClient *git.Client, desired *tree.Tree) func(mntr.Monitor) error {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing networking desired state")
		return PushGitDesiredStates(monitor, "Networking desired state written", gitClient, []GitDesiredState{{
			Desired: desired,
			Path:    git.NetworkingFile,
		}})
	}
}

type GitDesiredState struct {
	Desired *tree.Tree
	Path    git.DesiredFile
}

func PushGitDesiredStates(monitor mntr.Monitor, msg string, gitClient *git.Client, desireds []GitDesiredState) (err error) {
	monitor.OnChange = func(_ string, fields map[string]string) {
		gitFiles := make([]git.File, len(desireds))
		for i := range desireds {
			desired := desireds[i]
			gitFiles[i] = git.File{
				Path:    string(desired.Path),
				Content: common.MarshalYAML(desired.Desired),
			}
		}
		err = gitClient.UpdateRemote(mntr.SprintCommit(msg, fields), gitFiles...)
		mntr.LogMessage(msg, fields)
	}
	monitor.Changed(msg)
	return err
}
