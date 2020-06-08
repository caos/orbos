package kubectl

import (
	"strings"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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
		err = errors.Wrap(err, "Error while executing command")
		kubectlMonitor.Error(err)
		return "", err
	}

	versions := &Versions{}
	err = yaml.Unmarshal(out, versions)
	if err != nil {
		err = errors.Wrap(err, "Error while unmarshaling output")
		kubectlMonitor.Error(err)
		return "", err
	}

	parts := strings.Split(versions.ServerVersion.GitVersion, "-")
	return parts[0], nil
}
