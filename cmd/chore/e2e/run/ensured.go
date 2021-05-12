package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"

	"github.com/afiskon/promtail-client/promtail"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"gopkg.in/yaml.v3"
)

func ensureORBITERTest(logger promtail.Client, step uint8, timeout time.Duration, condition func(promtail.Client, newOrbctlCommandFunc, newKubectlCommandFunc) error) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc) error {

		triggerCheck := make(chan struct{})
		stopLogs := make(chan struct{})
		done := make(chan error)
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		go watchLogs(logger, kubectl, time.NewTimer(timeout), triggerCheck, stopLogs)

		started := time.Now()

		// Check each minute if the desired state is ensured
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-timer.C:
					done <- fmt.Errorf("timed out after %s", timeout)
				case <-triggerCheck:

					if err := condition(logger, orbctl, kubectl); err != nil {
						printProgress(logger, step, started, timeout)
						logger.Warnf("desired state is not yet ensured: %s", err.Error())
						continue
					}

					done <- nil
				case <-ticker.C:
					go func() { triggerCheck <- struct{}{} }()
				}
			}
		}()

		err := <-done
		stopLogs <- struct{}{}
		close(stopLogs)
		return err
	}
}

func watchLogs(logger promtail.Client, kubectl newKubectlCommandFunc, timer *time.Timer, trigger chan<- struct{}, stop <-chan struct{}) {
	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "logs", "--follow", "--selector", "app.kubernetes.io/name=orbiter", "--since", "10s")

	errWriter, errWrite := logWriter(logger.Warnf)
	defer errWrite()
	cmd.Stderr = errWriter

	goon := true
	go func() {
		<-stop
		goon = false
	}()

	err := simpleRunCommand(cmd, timer, func(line string) bool {
		// Check if the desired state is ensured when orbiter prints so
		if strings.Contains(line, "Desired state is ensured") {
			go func() { trigger <- struct{}{} }()
		}
		logORBITERStdout(logger, line)
		return goon
	})

	if !goon {
		return
	}

	if err != nil {
		logger.Warnf("watching logs failed: %s. trying again", err.Error())
	}

	watchLogs(logger, kubectl, timer, trigger, stop)
}

func isEnsured(orb string, masters, workers uint8, k8sVersion string) func(promtail.Client, newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(logger promtail.Client, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc) error {

		if err := checkPodsAreRunning(logger, newKubectl, "caos-system", "app.kubernetes.io/name=orbiter", 1); err != nil {
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
		if err := readYaml(logger, newOrbctl, "caos-internal/orbiter/current.yml", orbiter); err != nil {
			return err
		}

		nodeagents := &common.NodeAgentsCurrentKind{}
		if err := readYaml(logger, newOrbctl, "caos-internal/orbiter/node-agents-current.yml", nodeagents); err != nil {
			return err
		}

		cluster, ok := orbiter.Clusters[orb]
		if !ok {
			return fmt.Errorf("cluster %s not found in current state", orb)
		}
		currentMachinesLen := uint8(len(cluster.Current.Machines.M))

		machines := masters + workers

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
			if nodeagent.Software.Kubelet.Version != k8sVersion ||
				nodeagent.Software.Kubeadm.Version != k8sVersion ||
				nodeagent.Software.Kubectl.Version != k8sVersion {
				return fmt.Errorf("nodeagent %s has current states kubelet=%s, kubeadm=%s, kubectl=%s instead of %s",
					nodeagentID,
					nodeagent.Software.Kubelet.Version,
					nodeagent.Software.Kubeadm.Version,
					nodeagent.Software.Kubectl.Version,
					k8sVersion,
				)
			}
		}

		if cluster.Current.Status != "running" {
			return fmt.Errorf("cluster status is %s", cluster.Current.Status)
		}

		if err := checkPodsAreRunning(logger, newKubectl, "kube-system", "component in (etcd, kube-apiserver, kube-controller-manager, kube-scheduler)", masters*4); err != nil {
			return err
		}

		provider, ok := orbiter.Providers[orb]
		if !ok {
			return fmt.Errorf("provider %s not found in current state", orb)
		}

		ep := provider.Current.Ingresses.Httpsingress

		msg, err := helpers.Check("https", ep.Location, ep.FrontendPort, "/ambassador/v0/check_ready", 200, false)
		if err != nil {
			return fmt.Errorf("ambassador ready check failed with message: %s: %w", msg, err)
		}
		logger.Infof(msg)

		return nil
	}
}

func readYaml(logger promtail.Client, newOrbctl newOrbctlCommandFunc, path string, into interface{}) error {

	orbctl, err := newOrbctl()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	orbctl.Args = append(orbctl.Args, "--gitops", "file", "print", path)
	orbctl.Stdout = buf

	errWriter, errWrite := logWriter(logger.Warnf)
	defer errWrite()
	orbctl.Stderr = errWriter

	if err := orbctl.Run(); err != nil {
		return fmt.Errorf("reading orbiters current state failed: %w", err)
	}

	currentBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(currentBytes, into)
}

func checkPodsAreRunning(logger promtail.Client, kubectl newKubectlCommandFunc, namespace, selector string, expectedPodsCount uint8) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("check for running pods in namespace %s with selector %s failed: %w", namespace, selector, err)
		}
	}()

	buf := new(bytes.Buffer)
	defer buf.Reset()

	cmd := kubectl()
	cmd.Args = append(cmd.Args,
		"get", "pods",
		"--namespace", namespace,
		"--selector", selector,
		"--output", "yaml")
	cmd.Stdout = buf

	errWriter, errWrite := logWriter(logger.Warnf)
	defer errWrite()
	cmd.Stderr = errWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("getting container status failed: %w", err)
	}

	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return fmt.Errorf("reading container status failed: %w", err)
	}

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

	if err := yaml.Unmarshal(data, &pods); err != nil {
		return fmt.Errorf("unmarshalling container status failed: %w", err)
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
