package cli

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/orb"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

func PatchFile(
	orbConfig *orb.Orb,
	gitClient *git.Client,
	path []string,
	contentStr string,
	exact bool,
	filePath string,
) error {
	if err := InitRepo(orbConfig, gitClient); err != nil {
		return err
	}

	contentYaml := yamlTypedValue(contentStr)
	var result interface{}
	if len(path) == 0 && exact {
		result = contentYaml
	} else {
		structure := map[string]interface{}{}
		if err := yaml.Unmarshal(gitClient.Read(filePath), structure); err != nil {
			return err
		}
		if err := updateMap(structure, path, contentYaml, exact); err != nil {
			return err
		}
		result = structure
	}

	return gitClient.UpdateRemote(fmt.Sprintf("Overwrite %s", filePath), func() []git.File {
		return []git.File{{
			Path:    filePath,
			Content: common.MarshalYAML(result),
		}}
	})

}

func updateMap(structure map[string]interface{}, path []string, value interface{}, exact bool) error {

	if len(path) == 1 && exact {
		structure[path[0]] = value
		return nil
	}

	path, nextStep := drillPath(path)

	if nextStep == "" {

		var keys []string
		for key := range structure {
			keys = append(keys, key)
		}

		var err error
		nextStep, err = prompt(keys)
		if err != nil {
			return err
		}
	}

	child, ok := structure[nextStep]
	if !ok {
		return fmt.Errorf("path element %s not found", nextStep)
	}

	drilled, err := drillContent(child, path, value, exact)
	if err != nil {
		return err
	}
	if !drilled {
		structure[nextStep] = value
	}

	return nil
}

func prompt(keys []string) (string, error) {
	var key string
	return key, survey.AskOne(&survey.Select{
		Message: "Select key:",
		Options: keys,
	}, &key, survey.WithValidator(survey.Required))
}

func updateSlice(slice []interface{}, path []string, value interface{}, exact bool) error {

	pos := func(pathNode string) (int, error) {
		idx, err := strconv.Atoi(pathNode)
		if err != nil {
			return -1, err
		}

		length := len(slice)
		if length < idx {
			return -1, fmt.Errorf("property has only %d elements", length)
		}
		return idx, nil
	}

	if len(path) == 1 && exact {
		idx, err := pos(path[0])
		if err != nil {
			return err
		}

		slice[idx] = value
		return nil
	}

	path, nextStep := drillPath(path)

	if nextStep == "" {
		var keys []string
		for key := range slice {
			keys = append(keys, strconv.Itoa(key))
		}

		var err error
		nextStep, err = prompt(keys)
		if err != nil {
			return err
		}
	}

	idx, err := pos(nextStep)

	drilled, err := drillContent(slice[idx], path, value, exact)
	if err != nil {
		return err
	}
	if !drilled {
		slice[idx] = value
	}

	return nil
}

func drillPath(path []string) ([]string, string) {
	var next string

	if len(path) > 0 {
		next = path[0]
		path = path[1:]
	}
	return path, next
}

func drillContent(child interface{}, path []string, value interface{}, exact bool) (bool, error) {

	switch typedNext := child.(type) {
	case map[string]interface{}:
		return true, updateMap(typedNext, path, value, exact)
	case []interface{}:
		return true, updateSlice(typedNext, path, value, exact)
	}

	if len(path) > 0 {
		return false, fmt.Errorf("invalid path %s", strings.Join(path, "."))
	}

	return false, nil
}

func yamlTypedValue(value string) interface{} {

	var out interface{}
	if err := yaml.Unmarshal([]byte(value), &out); err != nil {
		panic(err)
	}
	return out
}
