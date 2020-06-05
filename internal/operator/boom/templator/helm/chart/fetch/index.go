package fetch

import (
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type index struct {
	APIVersion string             `yaml:"apiVersion"`
	Entries    map[string][]entry `yaml:"entries"`
}

type entry struct {
	Version    string `yaml:"version"`
	AppVersion string `yaml:"appVersion"`
}

func CompareVersions(monitor mntr.Monitor, basePath string, charts []*ChartInfo) error {

	indexFolderPathAbs, err := helper2.GetAbsPath(basePath, "helm", "repository", "cache")
	if err != nil {
		return err
	}

	indexFiles := make(map[string]*index, 0)
	err = filepath.Walk(indexFolderPathAbs, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var indexFile index
		if err := yaml.Unmarshal(data, &indexFile); err != nil {
			return err
		}

		indexFiles[strings.TrimSuffix(info.Name(), "-index.yaml")] = &indexFile
		// for k, v := range indexFile.Entries {
		// 	indexFiles[strings.TrimSuffix(info.Name(), "-index.yaml")].Entries[k] = v
		// }
		return nil
	})
	if err != nil {
		return err
	}
	for _, chart := range charts {
		indexFile := indexFiles[chart.IndexName]
		for _, entry := range indexFile.Entries[chart.Name] {
			entryParts := strings.Split(entry.Version, ".")
			entryPartsInt := make([]int, 3)
			for k, v := range entryParts {
				entryPartsInt[k], _ = strconv.Atoi(v)
			}

			chartParts := strings.Split(chart.Version, ".")
			chartPartsInt := make([]int, 3)
			for k, v := range chartParts {
				chartPartsInt[k], _ = strconv.Atoi(v)
			}
			if entryPartsInt[0] > chartPartsInt[0] ||
				(entryPartsInt[0] == chartPartsInt[0] && entryPartsInt[1] > chartPartsInt[1]) ||
				(entryPartsInt[0] == chartPartsInt[0] && entryPartsInt[1] == chartPartsInt[1] && entryPartsInt[2] > chartPartsInt[2]) {

				logFields := map[string]interface{}{
					"oldVersion":    chart.Version,
					"newVersion":    entry.Version,
					"index":         chart.IndexName,
					"chart":         chart.Name,
					"newAppVersion": entry.AppVersion,
				}
				monitor.WithFields(logFields).Info("There is a newer version")
			}
		}
	}

	return nil
}
