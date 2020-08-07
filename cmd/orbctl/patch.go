// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/common"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func PatchCommand(rv RootValues) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "patch <filepath> [yamlpath]",
		Short:   "Patch a yaml property",
		Args:    cobra.MinimumNArgs(1),
		Example: `orbctl file patch orbiter.yml`,
	}
	flags := cmd.Flags()
	var value string
	flags.StringVar(&value, "value", "", "The properties new value")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		_, _, orbConfig, gitClient := rv()

		if err := initRepo(orbConfig, gitClient); err != nil {
			return err
		}

		structure := map[string]interface{}{}
		if err := yaml.Unmarshal(gitClient.Read(args[0]), structure); err != nil {
			return err
		}

		var path []string
		if len(args) > 1 {
			path = strings.Split(args[1], ".")
		}

		if err := updateMap(structure, path, value); err != nil {
			return err
		}

		return gitClient.UpdateRemote(fmt.Sprintf("Overwrite %s", args[0]), git.File{
			Path:    args[0],
			Content: common.MarshalYAML(structure),
		})
	}

	return cmd
}

func updateMap(structure map[string]interface{}, path []string, value string) error {

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

	drilled, err := drillContent(child, path, value)
	if err != nil {
		return err
	}
	if !drilled {
		structure[nextStep] = yamlTypedValue(value)
	}

	return nil
}

func prompt(keys []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select key",
		Items: keys,
	}

	_, key, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return key, nil
}

func updateSlice(slice []interface{}, path []string, value string) error {

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

	idx, err := strconv.Atoi(nextStep)
	if err != nil {
		return err
	}

	length := len(slice)
	if length < idx {
		return fmt.Errorf("property has only %d elements", length)
	}

	drilled, err := drillContent(slice[idx-1], path, value)
	if err != nil {
		return err
	}
	if !drilled {
		slice[idx-1] = yamlTypedValue(value)
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

func drillContent(child interface{}, path []string, value string) (bool, error) {
	switch typedNext := child.(type) {
	case map[string]interface{}:
		return true, updateMap(typedNext, path, value)
	case []interface{}:
		return true, updateSlice(typedNext, path, value)
	}

	if len(path) > 0 {
		return false, fmt.Errorf("invalid path %s", strings.Join(path, "."))
	}

	return false, nil
}

func yamlTypedValue(value string) interface{} {
	v, err := strconv.Atoi(value)
	if err == nil {
		return v
	}

	return value
}
