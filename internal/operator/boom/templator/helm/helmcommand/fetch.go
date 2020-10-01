package helmcommand

import (
	"strings"

	helper2 "github.com/caos/orbos/internal/utils/helper"
)

type FetchConfig struct {
	TempFolderPath string
	ChartName      string
	ChartVersion   string
	IndexName      string
}

func FetchChart(conf *FetchConfig) error {

	chartHomeAbs, err := helper2.GetAbsPath(conf.TempFolderPath, chartsFolder)
	if err != nil {
		return err
	}

	versionParam := strings.Join([]string{"--version=", conf.ChartVersion}, "")
	untardirParam := strings.Join([]string{"--untardir=", chartHomeAbs}, "")
	chartStr := strings.Join([]string{conf.IndexName, conf.ChartName}, "/")
	command := strings.Join([]string{"fetch --untar", versionParam, untardirParam, chartStr}, " ")

	return doHelmCommand(conf.TempFolderPath, command)
}

type IndexConfig struct {
	TempFolderPath string
	IndexName      string
	IndexURL       string
}

func AddIndex(conf *IndexConfig) error {

	url := strings.Join([]string{"https://", conf.IndexURL}, "")
	command := strings.Join([]string{"repo add", conf.IndexName, url}, " ")

	return doHelmCommand(conf.TempFolderPath, command)
}

func RepoUpdate(basePath string) error {
	return doHelmCommand(basePath, "repo update")
}
