package kubectl

import "os/exec"

type KubectlDelete struct {
	kubectl *Kubectl
}

func NewDelete(resultFilePath string) *KubectlDelete {
	return &KubectlDelete{kubectl: New("delete").AddParameter("-f", resultFilePath).AddFlag("--ignore-not-found")}
}

func (k *KubectlDelete) Build() exec.Cmd {
	return k.kubectl.Build()
}
