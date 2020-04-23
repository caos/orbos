package push

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/tree"
	"github.com/caos/orbiter/mntr"
)

type Func func(monitor mntr.Monitor) error

func SecretsFunc(gitClient *git.Client, desired *tree.Tree, path string) Func {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing secret")
		return YML(monitor, "Secret written", gitClient, desired, path)
	}
}

func YML(monitor mntr.Monitor, msg string, gitClient *git.Client, desired *tree.Tree, path string) (err error) {
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
