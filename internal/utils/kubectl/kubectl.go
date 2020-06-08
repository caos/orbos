package kubectl

import (
	"os/exec"
)

type Kubectl struct {
	parameters []string
}

func New(command string) *Kubectl {
	return &Kubectl{
		parameters: []string{command},
	}
}

func (k *Kubectl) AddParameter(key, value string) *Kubectl {
	k.parameters = append(k.parameters, key)
	k.parameters = append(k.parameters, value)
	return k
}

func (k *Kubectl) AddFlag(flag string) *Kubectl {
	k.parameters = append(k.parameters, flag)
	return k
}

func (k *Kubectl) Build() exec.Cmd {
	cmd := exec.Command("kubectl", k.parameters...)
	return *cmd
}
