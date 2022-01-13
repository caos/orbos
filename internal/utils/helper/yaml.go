package helper

import (
	"strings"

	yamlfile "github.com/caos/orbos/internal/utils/yaml"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Resource struct {
	Kind       string    `yaml:"kind"`
	ApiVersion string    `yaml:"apiVersion"`
	Metadata   *Metadata `yaml:"metadata"`
}

func AddStringBeforePointForKindAndName(path, kind, name, point, addContent string) error {
	y := yamlfile.New(path)

	parts, err := y.ToObjectList()
	if err != nil {
		return err
	}

	if err := y.DeleteFile(); err != nil {
		return err
	}

	for _, part := range parts {
		if part == "" {
			continue
		}
		struc := &Resource{}
		if err := yaml.Unmarshal([]byte(part), struc); err != nil {
			return err
		}
		output := part
		if struc.Kind == kind && struc.Metadata.Name == name {
			lines := strings.Split(part, "\n")
			for i, line := range lines {
				if strings.Contains(line, point) {
					lines[i] = strings.Join([]string{addContent, line}, "")
				}
			}
			output = strings.Join(lines, "\n")
		}

		if err := y.AddStringObject(output); err != nil {
			return err
		}
	}
	return nil
}

func DeleteKindFromYaml(path, kind string) error {
	y := yamlfile.New(path)

	parts, err := y.ToObjectList()
	if err != nil {
		return err
	}

	if err := y.DeleteFile(); err != nil {
		return err
	}

	for _, part := range parts {

		struc := &Resource{}
		if err := yaml.Unmarshal([]byte(part), struc); err != nil {
			return err
		}

		if struc.Kind != kind {
			if err := y.AddStringObject(part); err != nil {
				return err
			}
		}
	}

	return nil
}

func DeleteFirstResourceFromYaml(path, apiVersion, kind, name string) error {
	y := yamlfile.New(path)

	parts, err := y.ToObjectList()
	if err != nil {
		return err
	}

	if err := y.DeleteFile(); err != nil {
		return err
	}

	found := false
	for _, part := range parts {
		struc := &Resource{}
		if err := yaml.Unmarshal([]byte(part), struc); err != nil {
			return err
		}

		if found || !(struc.ApiVersion == apiVersion && struc.Kind == kind && struc.Metadata.Name == name) {
			if err := y.AddStringObject(part); err != nil {
				return err
			}
		} else {
			found = true
		}
	}
	return nil
}
