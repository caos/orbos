package kubernetes

import (
	"errors"
	"strconv"
	"strings"

	"github.com/caos/orbiter/internal/operator/common"
)

type KubernetesVersion int

const (
	Unknown KubernetesVersion = iota
	V1x15x0
	V1x15x1
	V1x15x2
	V1x15x3
	V1x15x4
	V1x15x5
	V1x15x6
	V1x15x7
	V1x15x8
	V1x15x9
	V1x16x0
	V1x16x1
	V1x16x2
	V1x16x3
	V1x16x4
	V1x16x5
	V1x16x6
	V1x17x0
	V1x17x1
	V1x17x2
	V1x18x0
)

var kubernetesVersions = []string{
	"unknown",
	"v1.15.0", "v1.15.1", "v1.15.2", "v1.15.3", "v1.15.4", "v1.15.5", "v1.15.6", "v1.15.7", "v1.15.8", "v1.15.9",
	"v1.16.0", "v1.16.1", "v1.16.2", "v1.16.3", "v1.16.4", "v1.16.5", "v1.16.6",
	"v1.17.0", "v1.17.1", "v1.17.2",
	"v0.18.0"}

func (k KubernetesVersion) String() string {
	return kubernetesVersions[k]
}

func (k KubernetesVersion) DefineSoftware() common.Software {
	dockerVersion := "docker-ce v19.03.5"
	//	if minor, err := k.ExtractMinor(); err != nil && minor <= 15 {
	//		dockerVersion = "docker-ce v18.09.6"
	//	}
	return common.Software{
		Swap:             common.Package{Version: "disabled"},
		Containerruntime: common.Package{Version: dockerVersion},
		Kubelet:          common.Package{Version: k.String()},
		Kubeadm:          common.Package{Version: k.String()},
		Kubectl:          common.Package{Version: k.String()},
	}
}

func KubernetesSoftware(current common.Software) common.Software {
	return common.Software{
		Swap:             current.Swap,
		Containerruntime: current.Containerruntime,
		Kubelet:          current.Kubelet,
		Kubeadm:          current.Kubeadm,
		Kubectl:          current.Kubectl,
	}
}

func ParseString(version string) KubernetesVersion {
	for idx, k8sVersion := range kubernetesVersions {
		if k8sVersion == version {
			return KubernetesVersion(idx)
		}
	}
	return KubernetesVersion(0)
}

func (k KubernetesVersion) equals(other KubernetesVersion) bool {
	return string(k) == string(other)
}

func (k KubernetesVersion) NextHighestMinor() KubernetesVersion {
	switch k {
	case V1x15x0, V1x15x1, V1x15x2, V1x15x3, V1x15x4, V1x15x5, V1x15x6, V1x15x7, V1x15x8, V1x15x9:
		return V1x16x6
	case V1x16x0, V1x16x1, V1x16x2, V1x16x3, V1x16x4, V1x16x5, V1x16x6:
		return V1x17x2
	case V1x17x0, V1x17x1, V1x17x2:
		return V1x18x0
	default:
		return Unknown
	}
}

func (k KubernetesVersion) ExtractMinor() (int, error) {
	return k.extractNumber(1)
}

func (k KubernetesVersion) ExtractPatch() (int, error) {
	return k.extractNumber(2)
}

func (k KubernetesVersion) extractNumber(position int) (int, error) {
	if k == Unknown {
		return 0, errors.New("Unknown kubernetes version")
	}

	parts := strings.Split(k.String(), ".")
	version, err := strconv.ParseInt(parts[position], 10, 8)
	if err != nil {
		return 0, err
	}
	return int(version), nil
}
