package helmcommand

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	helper2 "github.com/caos/orbos/internal/utils/helper"
)

var (
	helmHomeFolder string = "helm"
	chartsFolder   string = "charts"
)

func Init(basePath string) error {
	return doHelmCommand(basePath, "init --client-only")
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

	stderr := new(bytes.Buffer)
	defer stderr.Reset()
	cmd := exec.Command("/bin/sh", "-c", helm)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while executing helm command \"%s\": %s: %w", helm, stderr.String(), err)
	}
	return nil
}

func doHelmCommandOutput(basePath, command string) ([]byte, error) {
	helmHomeFolderPathAbs, err := helper2.GetAbsPath(basePath, helmHomeFolder)
	if err != nil {
		return nil, err
	}

	helm := strings.Join([]string{"helm", "--home", helmHomeFolderPathAbs, command}, " ")

	cmd := exec.Command("/bin/sh", "-c", helm)
	return cmd.CombinedOutput()
}
