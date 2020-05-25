package helmcommand

import (
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var (
	helmHomeFolder string = "helm"
	chartsFolder   string = "charts"
)

func Init(basePath string) error {
	return doHelmCommand(basePath, "init --client-only >& /dev/null")
}

func addIfNotEmpty(one, two string) string {
	if two != "" {
		return strings.Join([]string{one, two}, " ")
	}
	return one
}

func doHelmCommand(basePath, command string) error {

	helmHomeFolderPathAbs, err := helper2.GetAbsPath(basePath, helmHomeFolder)
	if err != nil {
		return err
	}

	helm := strings.Join([]string{"helm", "--home", helmHomeFolderPathAbs, command}, " ")

	cmd := exec.Command("/bin/sh", "-c", helm)

	return errors.Wrapf(cmd.Run(), "Error while executing helm command \"%s\"", helm)
}

func doHelmCommandOutput(basePath, command string) ([]byte, error) {
	helmHomeFolderPathAbs, err := helper2.GetAbsPath(basePath, helmHomeFolder)
	if err != nil {
		return nil, err
	}

	helm := strings.Join([]string{"helm", "--home", helmHomeFolderPathAbs, command}, " ")

	cmd := exec.Command("/bin/sh", "-c", helm)
	return cmd.Output()
}
