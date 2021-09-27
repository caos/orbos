package desired

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yamlfile "github.com/caos/orbos/v5/internal/utils/yaml"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/internal/utils/helper"
	"github.com/caos/orbos/v5/internal/utils/kustomize"
	"github.com/caos/orbos/v5/mntr"
)

func Apply(monitor mntr.Monitor, resultFilePath, namespace string, appName name.Application, force bool) error {
	resultFileDirPath := filepath.Dir(resultFilePath)

	if err := prepareAdditionalFiles(resultFilePath, namespace, appName); err != nil {
		return err
	}

	// apply resources
	cmd, err := kustomize.New(resultFileDirPath)
	if err != nil {
		return err
	}
	cmd = cmd.Apply(force)
	err = helper.Run(monitor, cmd.Build())
	if err != nil {
		return fmt.Errorf("failed to apply with file %s: %w", resultFilePath, err)
	}

	return nil
}

func Get(monitor mntr.Monitor, resultFilePath, namespace string, appName name.Application) ([]*helper.Resource, error) {
	resultFileDirPath := filepath.Dir(resultFilePath)

	if err := prepareAdditionalFiles(resultFilePath, namespace, appName); err != nil {
		return nil, err
	}

	// apply resources
	cmd, err := kustomize.New(resultFileDirPath)
	if err != nil {
		return nil, err
	}

	out, err := helper.RunWithOutput(monitor, cmd.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to apply with file %s: %w", resultFilePath, err)
	}

	resources := make([]*helper.Resource, 0)

	parts := strings.Split(string(out), "\n---\n")
	for _, part := range parts {
		if part == "" {
			continue
		}
		var resource helper.Resource

		if err := yaml.Unmarshal([]byte(part), &resource); err != nil {
			return nil, err
		}

		resources = append(resources, &resource)
	}

	return resources, nil
}

func prepareAdditionalFiles(resultFilePath, namespace string, appName name.Application) error {
	resultFileDirPath := filepath.Dir(resultFilePath)

	resultFileKustomizePath := filepath.Join(resultFileDirPath, "kustomization.yaml")
	resultFileTransformerPath := filepath.Join(resultFileDirPath, "transformer.yaml")

	if helper.FileExists(resultFileKustomizePath) {
		if err := os.Remove(resultFileKustomizePath); err != nil {
			return err
		}
	}

	if helper.FileExists(resultFileTransformerPath) {
		if err := os.Remove(resultFileTransformerPath); err != nil {
			return err
		}
	}

	transformer := &kustomize.LabelTransformer{
		ApiVersion: "builtin",
		Kind:       "LabelTransformer",
		Metadata: &kustomize.Metadata{
			Name: "LabelTransformer",
		},
		Labels:     labels.GetAllApplicationLabels(appName),
		FieldSpecs: []*kustomize.FieldSpec{&kustomize.FieldSpec{Path: "metadata/labels", Create: true}},
	}
	if err := yamlfile.New(resultFileTransformerPath).AddStruct(transformer); err != nil {
		return err
	}

	kustomizeFile := kustomize.File{
		Namespace:    namespace,
		Resources:    []string{filepath.Base(resultFilePath)},
		Transformers: []string{filepath.Base(resultFileTransformerPath)},
	}

	if err := yamlfile.New(resultFileKustomizePath).AddStruct(kustomizeFile); err != nil {
		return err
	}
	return nil
}
