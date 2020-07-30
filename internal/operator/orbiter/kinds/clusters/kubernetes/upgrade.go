package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

type initializedMachines []*initializedMachine

func (c initializedMachines) Len() int           { return len(c) }
func (c initializedMachines) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c initializedMachines) Less(i, j int) bool { return c[i].infra.ID() < c[j].infra.ID() }

func ensureSoftware(
	monitor mntr.Monitor,
	target KubernetesVersion,
	k8sClient *Client,
	controlplane []*initializedMachine,
	workers []*initializedMachine) (bool, error) {

	sortedMachines := append(controlplane, workers...)
	from, to, err := findPath(monitor, sortedMachines, target)
	if err != nil {
		return false, err
	}

	monitor.WithFields(map[string]interface{}{
		"currentSoftware":   from,
		"currentKubernetes": from.Kubelet,
		"desiredSofware":    to,
		"desiredKubernetes": to.Kubelet,
	}).Debug("Ensuring kubernetes version")

	return step(k8sClient, monitor, sortedMachines, to)
}

func findPath(
	monitor mntr.Monitor,
	machines []*initializedMachine,
	target KubernetesVersion,
) (common.Software, common.Software, error) {

	var overallLowKubelet KubernetesVersion
	var overallLowKubeletMinor int
	zeroSW := common.Software{}

	for _, machine := range machines {
		id := machine.infra.ID()
		if machine.node == nil {
			continue
		}

		nodeinfoKubelet := machine.node.Status.NodeInfo.KubeletVersion

		monitor.WithFields(map[string]interface{}{
			"machine": id,
			"kubelet": nodeinfoKubelet,
		}).Debug("Found kubelet version from node info")
		kubelet := ParseString(nodeinfoKubelet)
		if kubelet == Unknown {
			return zeroSW, zeroSW, errors.Errorf("parsing version %s from nodes %s info failed", nodeinfoKubelet, id)
		}

		kubeletMinor, err := kubelet.ExtractMinor(monitor)
		if err != nil {
			return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from kubelet version %s from nodes %s info failed", nodeinfoKubelet, id)
		}

		if overallLowKubelet == Unknown {
			overallLowKubelet = kubelet
			overallLowKubeletMinor = kubeletMinor
			continue
		}

		kubeletPatch, err := kubelet.ExtractPatch(monitor)
		if err != nil {
			return zeroSW, zeroSW, errors.Wrapf(err, "extracting patch from kubelet version %s from nodes %s info failed", nodeinfoKubelet, id)
		}
		tmpOverallLowKubeletMinor, err := overallLowKubelet.ExtractMinor(monitor)
		if err != nil {
			return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from overall kubelet version %s failed", overallLowKubelet)
		}
		tmpOverallLowKubeletPatch, err := overallLowKubelet.ExtractPatch(monitor)
		if err != nil {
			return zeroSW, zeroSW, errors.Wrapf(err, "extracting patch from overall kubelet version %s failed", overallLowKubelet)
		}

		if kubeletMinor < tmpOverallLowKubeletMinor ||
			kubeletMinor == tmpOverallLowKubeletMinor && kubeletPatch < tmpOverallLowKubeletPatch {
			overallLowKubelet = kubelet
			overallLowKubeletMinor = kubeletMinor
		}
	}

	if overallLowKubelet == target || overallLowKubelet == Unknown {
		target := target.DefineSoftware()
		monitor.WithFields(map[string]interface{}{
			"from": overallLowKubelet,
			"to":   target,
		}).Debug("Cluster is up to date")
		return target, target, nil
	}

	targetMinor, err := target.ExtractMinor(monitor)
	if err != nil {
		return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from target version %s failed", target)
	}

	if targetMinor < overallLowKubeletMinor {
		return zeroSW, zeroSW, errors.Errorf("downgrading from %s to %s is not possible as they are on different minors", overallLowKubelet, target)
	}

	overallLowKubeletSoftware := overallLowKubelet.DefineSoftware()
	if (targetMinor - overallLowKubeletMinor) < 2 {
		monitor.WithFields(map[string]interface{}{
			"from":                   overallLowKubelet,
			"to":                     target,
			"targetMinor":            targetMinor,
			"overallLowKubeletMinor": overallLowKubeletMinor,
		}).Debug("Desired version can be reached directly")
		return overallLowKubeletSoftware, target.DefineSoftware(), nil
	}

	nextHighestMinor := overallLowKubelet.NextHighestMinor()
	monitor.WithFields(map[string]interface{}{
		"from":         overallLowKubelet,
		"fromMinor":    overallLowKubeletMinor,
		"intermediate": nextHighestMinor,
		"to":           target,
		"toMinor":      targetMinor,
	}).Debug("Desired version can be reached via an intermediate version")
	return overallLowKubeletSoftware, nextHighestMinor.DefineSoftware(), nil
}

func step(
	k8sClient *Client,
	monitor mntr.Monitor,
	sortedMachines initializedMachines,
	to common.Software,
) (bool, error) {

	for _, machine := range sortedMachines {
		if machine.node != nil && machine.node.Spec.Unschedulable && machine.node.Labels["orbos.ch/updating"] == machine.node.Status.NodeInfo.KubeletVersion {
			delete(machine.node.Labels, "orbos.ch/updating")
			if err := k8sClient.Uncordon(machine.currentMachine, machine.node); err != nil {
				return false, err
			}
		}
	}

	for idx, machine := range sortedMachines {

		next, err := plan(k8sClient, monitor, machine, idx == 0, to)
		if err != nil {
			return false, errors.Wrapf(err, "planning machine %s failed", machine.infra.ID())
		}

		if next == nil {
			continue
		}
		return false, next()
	}
	return true, nil
}

func plan(
	k8sClient *Client,
	monitor mntr.Monitor,
	machine *initializedMachine,
	isFirstControlplane bool,
	to common.Software,
) (func() error, error) {

	id := machine.infra.ID()
	machinemonitor := monitor.WithFields(map[string]interface{}{
		"machine": id,
	})

	awaitNodeAgent := func() error {
		machinemonitor.Info("Awaiting node agent")
		return nil
	}

	ensureSoftware := func(k8sNode *v1.Node, isControlplane bool, isFirstControlplane bool) func() error {
		return func() (err error) {

			defer func() {
				err = errors.Wrapf(err, "ensuring software on node %s failed", machine.infra.ID())
			}()

			if !isControlplane {
				k8sNode.Labels["orbos.ch/updating"] = to.Kubelet.Version
				if err := k8sClient.Drain(machine.currentMachine, k8sNode); err != nil {
					return err
				}
			}

			upgradeAction := "node"
			if isFirstControlplane {
				machinemonitor.Info("Migrating first controlplane node")
				upgradeAction = fmt.Sprintf("apply %s --yes", to.Kubelet.Version)
			} else {
				machinemonitor.Info("Migrating node")
			}

			_, err = machine.infra.Execute(nil, fmt.Sprintf("sudo kubeadm upgrade %s", upgradeAction))
			if err != nil {
				return err
			}

			if !softwareContains(*machine.desiredNodeagent.Software, to) {
				machine.desiredNodeagent.Software.Merge(to)
				machinemonitor.WithFields(map[string]interface{}{
					"from": machine.currentNodeagent.Software.Kubelet.Version,
					"to":   to.Kubelet.Version,
				}).Changed("Updated Kubernetes packages desired")
			}
			return nil
		}
	}

	ensureTargetSoftware := func(sw common.Software) func() error {
		return func() error {
			if !softwareContains(*machine.desiredNodeagent.Software, sw) {
				machine.desiredNodeagent.Software.Merge(to)
				machinemonitor.WithFields(map[string]interface{}{
					"current": KubernetesSoftware(machine.currentNodeagent.Software),
					"desired": sw,
				}).Changed("Software desired")
			} else {
				machinemonitor.Info("Awaiting kubernetes software")
			}
			return nil
		}
	}

	nodeIsReady := machine.currentNodeagent.NodeIsReady

	if !machine.currentMachine.Joined {
		if softwareContains(machine.currentNodeagent.Software, to) {
			if !nodeIsReady {
				return awaitNodeAgent, nil
			}

			// This node needs to be joined first
			return nil, nil
		}
		return ensureTargetSoftware(to), nil
	}

	alignedSoftware := new(common.Software)
	*alignedSoftware = machine.currentNodeagent.Software
	alignedSoftware.Kubeadm = common.Package{}
	if !softwareContains(*machine.desiredNodeagent.Software, *alignedSoftware) {
		return ensureTargetSoftware(*machine.desiredNodeagent.Software), nil
	}

	if machine.currentNodeagent.Software.Kubeadm.Version != to.Kubeadm.Version {
		return func() error {
			if machine.desiredNodeagent.Software.Kubeadm.Version != to.Kubeadm.Version {
				machine.desiredNodeagent.Software.Kubeadm.Version = to.Kubeadm.Version
				machinemonitor.WithFields(map[string]interface{}{
					"current": machine.currentNodeagent.Software.Kubeadm.Version,
					"desired": to.Kubeadm.Version,
				}).Changed("Kubeadm desired")
			}
			return nil
		}, nil
	}

	isControlplane := machine.pool.tier == Controlplane
	if machine.node.Status.NodeInfo.KubeletVersion != to.Kubelet.Version {
		return ensureSoftware(machine.node, isControlplane, isFirstControlplane), nil
	}

	if !softwareContains(machine.currentNodeagent.Software, to, true) || !softwareContains(*machine.desiredNodeagent.Software, to, true) {
		return ensureTargetSoftware(to), nil
	}

	if !nodeIsReady {
		return awaitNodeAgent, nil
	}

	return nil, nil
}
