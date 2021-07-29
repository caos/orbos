package kubectl

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/mntr"
)

type KubectlLabel struct {
	kubectl *Kubectl
}

func NewLabel(resultFilePath string) *KubectlLabel {
	return &KubectlLabel{kubectl: New("label").AddFlag("--overwrite").AddParameter("-f", resultFilePath)}
}
func (k *KubectlLabel) AddParameter(key, value string) *KubectlLabel {
	k.kubectl.AddParameter(key, value)
	return k
}

func (k *KubectlLabel) Apply(monitor mntr.Monitor, labels map[string]string) error {
	for key, value := range labels {
		k.kubectl.AddFlag(strings.Join([]string{key, value}, "="))
	}

	cmd := k.kubectl.Build()

	kubectlMonitor := monitor.WithFields(map[string]interface{}{
		"cmd": cmd,
	})
	kubectlMonitor.Debug("Executing")

	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error while executing command: %w", err)
		kubectlMonitor.Debug(string(out))
		kubectlMonitor.Error(err)
	}

	return err
}
