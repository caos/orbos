package k8s

import (
	"errors"
	"strconv"
	"strings"

	"github.com/caos/orbiter/internal/core/operator"
)

type KubernetesVersion int

const (
	Unknown KubernetesVersion = iota
	V1x15x0
	V1x15x1
	V1x15x2
	V1x15x3
	V1x15x4
	V1x16x0
	V1x16x1
	V1x16x2
	V1x16x3
	V1x16x4
	V1x17x0
)

var kubernetesVersions = []string{
	"unknown",
	"v1.15.0", "v1.15.1", "v1.15.2", "v1.15.3", "v1.15.4",
	"v1.16.0", "v1.16.1", "v1.16.2", "v1.16.3", "v1.16.4",
	"v1.17.0"}

func (k KubernetesVersion) String() string {
	return kubernetesVersions[k]
}

func (k KubernetesVersion) DefineSoftware() operator.Software {
	return operator.Software{
		Swap:             operator.Package{Version: "disabled"},
		Containerruntime: operator.Package{Version: "docker-ce v18.09.6"},
		Kubelet:          operator.Package{Version: k.String()},
		Kubeadm:          operator.Package{Version: k.String()},
		Kubectl:          operator.Package{Version: k.String()},
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
	case V1x15x0:
		fallthrough
	case V1x15x1:
		fallthrough
	case V1x15x2:
		fallthrough
	case V1x15x3:
		return V1x16x0
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
