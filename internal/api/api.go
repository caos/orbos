package api

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

const (
	orbiterFile = "orbiter.yml"
	boomFile    = "boom.yml"
)

type SecretFunc func(monitor mntr.Monitor) error

func ExistsOrbiterYml(gitClient *git.Client) (bool, error) {
	return existsFileInGit(gitClient, orbiterFile)
}

func ReadOrbiterYml(gitClient *git.Client) (*tree.Tree, error) {
	return readFileInGit(gitClient, orbiterFile)
}

func PushOrbiterYml(monitor mntr.Monitor, msg string, gitClient *git.Client, desired *tree.Tree) (err error) {
	return pushFileInGit(monitor, msg, gitClient, desired, orbiterFile)
}

func OrbiterSecretFunc(gitClient *git.Client, desired *tree.Tree) SecretFunc {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing orbiter secrets")
		return PushOrbiterYml(monitor, "Orbiter secrets written", gitClient, desired)
	}
}

func ExistsBoomYml(gitClient *git.Client) (bool, error) {
	return existsFileInGit(gitClient, boomFile)
}

func ReadBoomYml(gitClient *git.Client) (*tree.Tree, error) {
	return readFileInGit(gitClient, boomFile)
}

func PushBoomYml(monitor mntr.Monitor, msg string, gitClient *git.Client, desired *tree.Tree) (err error) {
	return pushFileInGit(monitor, msg, gitClient, desired, boomFile)
}

func BoomSecretFunc(gitClient *git.Client, desired *tree.Tree) SecretFunc {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing boom secrets")
		return PushBoomYml(monitor, "Orbiter secrets written", gitClient, desired)
	}
}

func pushFileInGit(monitor mntr.Monitor, msg string, gitClient *git.Client, desired *tree.Tree, path string) (err error) {
	monitor.OnChange = func(_ string, fields map[string]string) {
		err = gitClient.UpdateRemote(mntr.SprintCommit(msg, fields), git.File{
			Path:    path,
			Content: common.MarshalYAML(desired),
		})
		mntr.LogMessage(msg, fields)
	}
	monitor.Changed(msg)
	return err
}

func existsFileInGit(gitClient *git.Client, path string) (bool, error) {
	if err := gitClient.Clone(); err != nil {
		return false, err
	}

	of := gitClient.Read(path)
	if of != nil && len(of) > 0 {
		return true, nil
	}
	return false, nil
}

func readFileInGit(gitClient *git.Client, path string) (*tree.Tree, error) {
	if err := gitClient.Clone(); err != nil {
		return nil, err
	}

	tree := &tree.Tree{}
	if err := yaml.Unmarshal(gitClient.Read(path), tree); err != nil {
		return nil, err
	}

	return tree, nil
}
