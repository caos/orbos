package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"gopkg.in/yaml.v3"
)

func awaitORBITER(
	settings programSettings,
	timeout time.Duration,
	orbctl newOrbctlCommandFunc,
	kubectl newKubectlCommandFunc,
	downloadKubeconfigFunc downloadKubeconfig,
	step uint8,
	desired *kubernetes.Spec,
	furtherCurrentChecks checkCurrentFunc,
) error {

	ensureCtx, ensureCancel := context.WithTimeout(settings.ctx, timeout)
	defer ensureCancel()

	if err := downloadKubeconfigFunc(orbctl); err != nil {
		return err
	}

	triggerCheck := make(chan struct{})
	done := make(chan error)

	go watchLogs(ensureCtx, settings, kubectl, triggerCheck)

	started := time.Now()

	// Check each minute if the desired state is ensured
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ensureCtx.Done():
				done <- ensureCtx.Err()
				return
			case <-triggerCheck:

				if err := isEnsured(ensureCtx, settings, orbctl, kubectl, desired, furtherCurrentChecks); err != nil {
					printProgress(settings, strconv.Itoa(int(step)), started, timeout)
					settings.logger.Warnf("desired state is not yet ensured: %s", err.Error())
					continue
				}

				done <- nil
				return
			case <-ticker.C:
				go func() { triggerCheck <- struct{}{} }()
			}
		}
	}()

	return <-done
}

func watchLogs(ctx context.Context, settings programSettings, kubectl newKubectlCommandFunc, trigger chan<- struct{}) {

	select {
	case <-ctx.Done():
		return
	default:
		// goon
	}

	err := runCommand(settings, true, nil, func(line string) {
		// Check if the desired state is ensured when orbiter prints so
		if strings.Contains(line, "Desired state is ensured") {
			go func() { trigger <- struct{}{} }()
		}
	}, kubectl(ctx), "logs", "--namespace", "caos-system", "--selector", "app.kubernetes.io/name=orbiter", "--since", "10s", "--follow")

	if err != nil {
		settings.logger.Warnf("watching logs failed: %s. trying again", err.Error())
	}

	time.Sleep(1 * time.Second)

	watchLogs(ctx, settings, kubectl, trigger)
}

type checkCurrentFunc func(current common.NodeAgentsCurrentKind) error

func isEnsured(ctx context.Context, settings programSettings, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc, desired *kubernetes.Spec, furtherChecks checkCurrentFunc) error {

	if err := checkPodsAreRunning(ctx, settings, newKubectl, "caos-system", "app.kubernetes.io/name=orbiter", 1); err != nil {
		return err
	}

	orbiter := &struct {
		Clusters map[string]struct {
			Current kubernetes.CurrentCluster
		}
		Providers map[string]struct {
			Current struct {
				Ingresses struct {
					Httpsingress infra.Address
					Httpingress  infra.Address
					Kubeapi      infra.Address
				}
			}
		}
	}{}
	if err := readYaml(ctx, settings, newOrbctl, orbiter, "--gitops", "file", "print", "caos-internal/orbiter/current.yml"); err != nil {
		return err
	}

	nodeagents := &common.NodeAgentsCurrentKind{}
	if err := readYaml(ctx, settings, newOrbctl, nodeagents, "--gitops", "file", "print", "caos-internal/orbiter/node-agents-current.yml"); err != nil {
		return err
	}

	cluster, ok := orbiter.Clusters[settings.orbID]
	if !ok {
		return fmt.Errorf("cluster %s not found in current state", settings.orbID)
	}
	currentMachinesLen := uint8(len(cluster.Current.Machines.M))

	machines := uint8(desired.ControlPlane.Nodes)
	for i := range desired.Workers {
		machines += uint8(desired.Workers[i].Nodes)
	}

	if currentMachinesLen != machines {
		return fmt.Errorf("current state has %d machines instead of %d", currentMachinesLen, machines)
	}

	for nodeagentID, nodeagent := range cluster.Current.Machines.M {
		if !nodeagent.Ready ||
			!nodeagent.FirewallIsReady ||
			!nodeagent.Joined {
			return fmt.Errorf("nodeagent %s has current states joined=%t, firewallIsReady=%t, ready=%t",
				nodeagentID,
				nodeagent.Ready,
				nodeagent.FirewallIsReady,
				nodeagent.Joined,
			)
		}
	}

	for nodeagentID, nodeagent := range nodeagents.Current.NA {
		if !nodeagent.NodeIsReady {
			return fmt.Errorf("nodeagent %s has not reported readiness yet", nodeagentID)
		}
		if nodeagent.Software.Kubelet.Version != desired.Versions.Kubernetes ||
			nodeagent.Software.Kubeadm.Version != desired.Versions.Kubernetes ||
			nodeagent.Software.Kubectl.Version != desired.Versions.Kubernetes {
			return fmt.Errorf("nodeagent %s has current states kubelet=%s, kubeadm=%s, kubectl=%s instead of %s",
				nodeagentID,
				nodeagent.Software.Kubelet.Version,
				nodeagent.Software.Kubeadm.Version,
				nodeagent.Software.Kubectl.Version,
				desired.Versions.Kubernetes,
			)
		}
	}

	if cluster.Current.Status != "running" {
		return fmt.Errorf("cluster status is %s", cluster.Current.Status)
	}

	if err := checkPodsAreRunning(ctx, settings, newKubectl, "kube-system", "component in (etcd, kube-apiserver, kube-controller-manager, kube-scheduler)", uint8(desired.ControlPlane.Nodes*4)); err != nil {
		return err
	}

	provider, ok := orbiter.Providers[settings.orbID]
	if !ok {
		return fmt.Errorf("provider %s not found in current state", settings.orbID)
	}

	ep := provider.Current.Ingresses.Httpsingress

	msg, err := helpers.Check("https", ep.Location, ep.FrontendPort, "/ambassador/v0/check_ready", 200, false)
	if err != nil {
		return fmt.Errorf("ambassador ready check failed with message: %s: %w", msg, err)
	}

	if furtherChecks != nil {
		if err := furtherChecks(*nodeagents); err != nil {
			return err
		}
	}

	settings.logger.Infof(msg)

	return nil
}

func readYaml(ctx context.Context, settings programSettings, binFunc func(context.Context) *exec.Cmd, into interface{}, args ...string) error {

	orbctlCtx, orbctlCancel := context.WithTimeout(ctx, 15*time.Second)
	defer orbctlCancel()

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := runCommand(settings, false, buf, nil, binFunc(orbctlCtx), args...); err != nil {
		return fmt.Errorf("reading orbiters current state failed: %w", err)
	}

	currentBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(currentBytes, into)
}

func checkPodsAreRunning(ctx context.Context, settings programSettings, kubectl newKubectlCommandFunc, namespace, selector string, expectedPodsCount uint8) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("check for running pods in namespace %s with selector %s failed: %w", namespace, selector, err)
		}
	}()

	pods := struct {
		Items []struct {
			Metadata struct {
				Name string
			}
			Status struct {
				Conditions []struct {
					Type   string
					Status string
				}
			}
		}
	}{}

	if err := readYaml(ctx, settings, kubectl, &pods, "get", "pods", "--namespace", namespace, "--selector", selector, "--output", "yaml"); err != nil {
		return err
	}

	podsCount := uint8(len(pods.Items))
	if podsCount != expectedPodsCount {
		return fmt.Errorf("%d pods are running instead of %d", podsCount, expectedPodsCount)
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		isReady := false
		for j := range pod.Status.Conditions {
			condition := pod.Status.Conditions[j]
			if condition.Type != "Ready" {
				continue
			}
			if condition.Status != "True" {
				return fmt.Errorf("pod %s has Ready condition %s", pod.Metadata.Name, condition.Status)
			}
			isReady = true
			break
		}
		if !isReady {
			return fmt.Errorf("pod %s has no Ready condition", pod.Metadata.Name)
		}
	}

	return nil
}
