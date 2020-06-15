package kustomize

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type File struct {
	Namespace    string   `yaml:"namespace"`
	Transformers []string `yaml:"transformers,omitempty"`
	Resources    []string `yaml:"resources"`
}
type LabelTransformer struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   *Metadata         `yaml:"metadata"`
	Labels     map[string]string `yaml:"labels"`
	FieldSpecs []*FieldSpec      `yaml:"fieldSpecs"`
}
type Metadata struct {
	Name string `yaml:"name"`
}
type FieldSpec struct {
	Kind   string `yaml:"kind,omitempty"`
	Path   string `yaml:"path"`
	Create bool   `yaml:"create"`
}

type Kustomize struct {
	path   string
	apply  bool
	delete bool
	force  bool
}

func New(path string) (*Kustomize, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return &Kustomize{
		path:   abspath,
		apply:  false,
		delete: false,
		force:  false,
	}, nil
}

func (k *Kustomize) Apply(force bool) *Kustomize {
	k.apply = true
	k.delete = false
	k.force = force
	return k
}
func (k *Kustomize) Delete() *Kustomize {
	k.apply = false
	k.delete = true
	return k
}

func (k *Kustomize) Build() exec.Cmd {
	all := strings.Join([]string{"kustomize", "build", k.path}, " ")
	if k.apply {
		all = strings.Join([]string{all, "| kubectl apply -f -"}, " ")
		if k.force {
			all = strings.Join([]string{all, "--force"}, " ")
		}
	} else if k.delete {
		all = strings.Join([]string{all, "| kubectl delete -f -"}, " ")
	}

	cmd := exec.Command("/bin/sh", "-c", all)
	return *cmd
}
