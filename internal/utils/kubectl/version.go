package kubectl

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/v5/mntr"
)

type Version struct {
	BuildDate    string `yaml:"buildDate"`
	Compiler     string `yaml:"compiler"`
	GitCommit    string `yaml:"gitCommit"`
	GitTreeState string `yaml:"gitTreeState"`
	GitVersion   string `yaml:"gitVersion"`
	GoVersion    string `yaml:"goVersion"`
	Major        string `yaml:"major"`
	Minor        string `yaml:"minor"`
	Platform     string `yaml:"platform"`
}

type Versions struct {
	ClientVersion *Version `yaml:"clientVersion"`
	ServerVersion *Version `yaml:"serverVersion"`
}

type KubectlVersion struct {
	kubectl *Kubectl
}

func NewVersion() *KubectlVersion {
	return &KubectlVersion{kubectl: New("version").AddParameter("-o", "yaml")}
}

func (k *KubectlVersion) GetKubeVersion(monitor mntr.Monitor) (string, error) {
	cmd := k.kubectl.Build()

	kubectlMonitor := monitor.WithFields(map[string]interface{}{
		"cmd": cmd,
	})
	kubectlMonitor.Debug("Executing")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error while executing command: %w", err)
	}

	versions := &Versions{}
	err = yaml.Unmarshal(out, versions)
	if err != nil {
		return "", fmt.Errorf("error while unmarshaling output: %w", err)
	}

	parts := strings.Split(versions.ServerVersion.GitVersion, "-")
	return parts[0], nil
}
