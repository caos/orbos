package kubernetes

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
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

	findPath := func(machines []*initializedMachine) (common.Software, common.Software, error) {

		var overallLowKubelet KubernetesVersion
		var overallLowKubeletMinor int
		zeroSW := common.Software{}

		for _, machine := range machines {
			id := machine.infra.ID()
			node, err := k8sClient.GetNode(id)
			if err != nil {
				continue
			}

			nodeinfoKubelet := node.Status.NodeInfo.KubeletVersion

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

	plan := func(
		machine *initializedMachine,
		isFirstControlplane bool,
		to common.Software) (func() error, error) {

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

				_, err = machine.infra.Execute(nil, nil, fmt.Sprintf("sudo kubeadm upgrade %s", upgradeAction))
				if err != nil {
					return err
				}

				if !machine.desiredNodeagent.Software.Contains(to) {
					machine.desiredNodeagent.Software.Merge(to)
					machinemonitor.WithFields(map[string]interface{}{
						"from": machine.currentNodeagent.Software.Kubelet.Version,
						"to":   to.Kubelet.Version,
					}).Changed("Updated Kubernetes packages desired")
				}
				return nil
			}
		}

		ensureOnline := func(k8sNode *v1.Node) func() error {
			return func() error {
				if err := k8sClient.Uncordon(machine.currentMachine, k8sNode); err != nil {
					return err
				}
				return nil
			}
		}

		if !machine.currentNodeagent.NodeIsReady {
			return awaitNodeAgent, nil
		}

		if !machine.currentMachine.Joined {
			if machine.currentNodeagent.Software.Contains(to) {
				// This node needs to be joined first
				return nil, nil
			}
			return func() error {
				if !machine.desiredNodeagent.Software.Contains(to) {
					machine.desiredNodeagent.Software.Merge(to)
					machinemonitor.WithFields(map[string]interface{}{
						"current": KubernetesSoftware(machine.currentNodeagent.Software),
						"desired": to,
					}).Changed("Join software desired")
				}
				return nil
			}, nil
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

		isControlplane := machine.tier == Controlplane
		k8sNode, err := k8sClient.GetNode(id)
		if err != nil {
			return nil, err
		}

		if k8sNode.Status.NodeInfo.KubeletVersion != to.Kubelet.Version {
			return ensureSoftware(k8sNode, isControlplane, isFirstControlplane), nil
		}

		if k8sNode.Spec.Unschedulable && !isControlplane {
			machine.currentMachine.Online = false
			return ensureOnline(k8sNode), nil
		}

		if !machine.currentNodeagent.NodeIsReady || !machine.currentNodeagent.Software.Contains(to) {
			return awaitNodeAgent, nil
		}

		return nil, nil
	}

	sortedControlplane := initializedMachines(controlplane)
	sortedWorkers := initializedMachines(workers)
	sort.Sort(sortedControlplane)
	sort.Sort(sortedWorkers)

	from, to, err := findPath(append(controlplane, workers...))
	if err != nil {
		return false, err
	}

	monitor.WithFields(map[string]interface{}{
		"currentSoftware":   from,
		"currentKubernetes": from.Kubelet,
		"desiredSofware":    to,
		"desiredKubernetes": to.Kubelet,
	}).Debug("Ensuring kubernetes version")

	done := true
	nexting := true
	for idx, machine := range append(sortedControlplane, sortedWorkers...) {

		next, err := plan(machine, idx == 0, to)
		if err != nil {
			return false, errors.Wrapf(err, "planning machine %s failed", machine.infra.ID())
		}

		if next == nil || !nexting {
			continue
		}

		done = false

		if err := next(); err != nil {
			return false, err
		}
		nexting = false
	}

	return done, nil
}
