package kubernetes

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
)

func join(
	joining infra.Compute,
	joinAt infra.Compute,
	cfg *model.Config,
	kubeAPI infra.Address,
	joinToken string,
	kubernetesVersion k8s.KubernetesVersion,
	certKey string,
	isControlPlane bool) (*string, error) {

	var installNetwork func() error
	switch cfg.Spec.Networking.Network {
	case "cilium":
		installNetwork = func() error {
			return try(cfg.Params.Logger, time.NewTimer(20*time.Second), 2*time.Second, joining, func(cmp infra.Compute) error {
				applyStdout, applyErr := cmp.Execute(nil, nil, "kubectl create -f https://raw.githubusercontent.com/cilium/cilium/1.6.3/install/kubernetes/quick-install.yaml")
				cfg.Params.Logger.WithFields(map[string]interface{}{
					"stdout": string(applyStdout),
				}).Debug("Applied cilium network")
				return applyErr
			})
		}
	case "calico":
		installNetwork = func() error {
			return try(cfg.Params.Logger, time.NewTimer(20*time.Second), 2*time.Second, joining, func(cmp infra.Compute) error {
				applyStdout, applyErr := cmp.Execute(nil, nil, fmt.Sprintf(`curl https://docs.projectcalico.org/v3.10/manifests/calico.yaml -O && sed -i -e "s?192.168.0.0/16?%s?g" calico.yaml && kubectl apply -f calico.yaml`, cfg.Spec.Networking.PodCidr))
				cfg.Params.Logger.WithFields(map[string]interface{}{
					"stdout": string(applyStdout),
				}).Debug("Applied calico network")
				return applyErr
			})
		}
	default:
		return nil, errors.Errorf("Unknown network implementation %s", cfg.Spec.Networking.Network)
	}

	intIP := joining.IP()

	kubeadmCfgPath := "/etc/kubeadm/config.yaml"
	kubeadmCfg := fmt.Sprintf(`apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: %s
  ttl: 10m0s
  usages:
  - signing
  - authentication
localAPIEndpoint:
  advertiseAddress: %s
  bindPort: 6666
nodeRegistration:
#	criSocket: /var/run/dockershim.sock
  name: %s
  taints:
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
cgroupDriver: systemd
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
apiServer:
  timeoutForControlPlane: 4m0s
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controlPlaneEndpoint: %s
controllerManager: {}
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: k8s.gcr.io
kubernetesVersion: %s
networking:
  dnsDomain: %s
  podSubnet: %s
  serviceSubnet: %s
scheduler: {}
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: JoinConfiguration
caCertPath: /etc/kubernetes/pki/ca.crt
discovery:
  bootstrapToken:
    apiServerEndpoint: %s
    token: %s
    unsafeSkipCAVerification: true
  timeout: 5m0s
nodeRegistration:
  name: %s
`,
		joinToken,
		intIP,
		joining.ID(),
		kubeAPI,
		kubernetesVersion,
		cfg.Spec.Networking.DNSDomain,
		cfg.Spec.Networking.PodCidr,
		cfg.Spec.Networking.ServiceCidr,
		kubeAPI,
		joinToken,
		joining.ID())

	if isControlPlane {
		kubeadmCfg += fmt.Sprintf(`controlPlane:
  localAPIEndpoint:
    advertiseAddress: %s
    bindPort: 6666
  certificateKey: %s
`, intIP, certKey)
	}

	if err := try(cfg.Params.Logger, time.NewTimer(7*time.Second), 2*time.Second, joining, func(cmp infra.Compute) error {
		return cmp.WriteFile(kubeadmCfgPath, strings.NewReader(kubeadmCfg), 600)
	}); err != nil {
		return nil, err
	}
	cfg.Params.Logger.WithFields(map[string]interface{}{
		"path": kubeadmCfgPath,
	}).Debug("Written file")

	cmd := fmt.Sprintf("sudo kubeadm reset -f && sudo rm -rf /var/lib/etcd")
	resetStdout, err := joining.Execute(nil, nil, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "executing %s failed", cmd)
	}
	cfg.Params.Logger.WithFields(map[string]interface{}{
		"stdout": string(resetStdout),
	}).Debug("Cleaned up compute")

	if joinAt != nil {
		joinAtIP := joinAt.IP()
		if err != nil {
			return nil, err
		}

		cmd := fmt.Sprintf("sudo kubeadm join --ignore-preflight-errors=Port-%d %s:%d --config %s", kubeAPI.Port, joinAtIP, kubeAPI.Port, kubeadmCfgPath)
		joinStdout, err := joining.Execute(nil, nil, cmd)
		if err != nil {
			return nil, errors.Wrapf(err, "executing %s failed", cmd)
		}
		cfg.Params.Logger.WithFields(map[string]interface{}{
			"stdout": string(joinStdout),
		}).Debug("Executed kubeadm join")
		return nil, nil
	}

	var kubeconfig bytes.Buffer
	initCmd := fmt.Sprintf("sudo kubeadm init --ignore-preflight-errors=Port-%d --config %s", kubeAPI.Port, kubeadmCfgPath)
	initStdout, err := joining.Execute(nil, nil, initCmd)
	if err != nil {
		return nil, err
	}
	cfg.Params.Logger.WithFields(map[string]interface{}{
		"stdout": string(initStdout),
	}).Debug("Executed kubeadm init")

	copyKubeconfigStdout, err := joining.Execute(nil, nil, fmt.Sprintf("mkdir -p ${HOME}/.kube && yes | sudo cp -rf /etc/kubernetes/admin.conf ${HOME}/.kube/config && sudo chown $(id -u):$(id -g) ${HOME}/.kube/config"))
	cfg.Params.Logger.WithFields(map[string]interface{}{
		"stdout": string(copyKubeconfigStdout),
	}).Debug("Moved kubeconfig")
	if err != nil {
		return nil, err
	}

	if err := installNetwork(); err != nil {
		return nil, err
	}

	if err := joining.ReadFile("${HOME}/.kube/config", &kubeconfig); err != nil {
		return nil, err
	}

	kc := kubeconfig.String()

	return &kc, nil
}
