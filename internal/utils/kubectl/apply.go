package kubectl

import "os/exec"

type KubectlApply struct {
	kubectl *Kubectl
}

func NewApply(resultFilePath string) *KubectlApply {
	return &KubectlApply{kubectl: New("apply").AddParameter("-f", resultFilePath)}
}

func (k *KubectlApply) Force() *KubectlApply {
	k.kubectl.AddFlag("--force")
	return k
}

func (k *KubectlApply) Build() exec.Cmd {
	return k.kubectl.Build()
}
