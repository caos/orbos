package kubernetes

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

type initializedMachines []initializedMachine

func (c initializedMachines) Len() int           { return len(c) }
func (c initializedMachines) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c initializedMachines) Less(i, j int) bool { return c[i].infra.ID() < c[j].infra.ID() }

func ensureSoftware(
	logger logging.Logger,
	target k8s.KubernetesVersion,
	k8sClient *k8s.Client,
	controlplane []initializedMachine,
	workers []initializedMachine) (bool, error) {

	findPath := func(machines []initializedMachine) (common.Software, common.Software, error) {

		var overallLowKubelet k8s.KubernetesVersion
		var overallLowKubeletMinor int
		zeroSW := common.Software{}

		for _, machine := range machines {
			id := machine.infra.ID()
			node, err := k8sClient.GetNode(id)
			if err != nil {
				continue
			}

			nodeinfoKubelet := node.Status.NodeInfo.KubeletVersion

			logger.WithFields(map[string]interface{}{
				"machine": id,
				"kubelet": nodeinfoKubelet,
			}).Debug("Found kubelet version from node info")
			kubelet := k8s.ParseString(nodeinfoKubelet)
			if kubelet == k8s.Unknown {
				return zeroSW, zeroSW, errors.Errorf("parsing version %s from nodes %s info failed", nodeinfoKubelet, id)
			}

			kubeletMinor, err := kubelet.ExtractMinor()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from kubelet version %s from nodes %s info failed", nodeinfoKubelet, id)
			}

			if overallLowKubelet == k8s.Unknown {
				overallLowKubelet = kubelet
				overallLowKubeletMinor = kubeletMinor
				continue
			}

			kubeletPatch, err := kubelet.ExtractPatch()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting patch from kubelet version %s from nodes %s info failed", nodeinfoKubelet, id)
			}
			overallLowKubeletMinor, err := overallLowKubelet.ExtractMinor()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from overall kubelet version %s failed", overallLowKubelet)
			}
			overallLowKubeletPatch, err := overallLowKubelet.ExtractPatch()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting patch from overall kubelet version %s failed", overallLowKubelet)
			}

			if kubeletMinor < overallLowKubeletMinor ||
				kubeletMinor == overallLowKubeletMinor && kubeletPatch < overallLowKubeletPatch {
				overallLowKubelet = kubelet
				overallLowKubeletMinor = kubeletMinor
			}
		}

		if overallLowKubelet == target || overallLowKubelet == k8s.Unknown {
			target := target.DefineSoftware()
			logger.WithFields(map[string]interface{}{
				"from": overallLowKubelet,
				"to":   target,
			}).Debug("Cluster is up to date")
			return target, target, nil
		}

		targetMinor, err := target.ExtractMinor()
		if err != nil {
			return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from target version %s failed", target)
		}

		if targetMinor < overallLowKubeletMinor {
			return zeroSW, zeroSW, errors.Errorf("downgrading from %s to %s is not possible as they are on different minors", overallLowKubelet, target)
		}

		overallLowKubeletSoftware := overallLowKubelet.DefineSoftware()
		if targetMinor-overallLowKubeletMinor < 2 {
			logger.WithFields(map[string]interface{}{
				"from": overallLowKubelet,
				"to":   target,
			}).Debug("Desired version can be reached directly")
			return overallLowKubeletSoftware, target.DefineSoftware(), nil
		}

		nextHighestMinor := overallLowKubelet.NextHighestMinor()
		logger.WithFields(map[string]interface{}{
			"from":         overallLowKubelet,
			"fromMinor":    overallLowKubeletMinor,
			"intermediate": nextHighestMinor,
			"to":           target,
			"toMinor":      targetMinor,
		}).Debug("Desired version can be reached via an intermediate version")
		return overallLowKubeletSoftware, nextHighestMinor.DefineSoftware(), nil
	}

	plan := func(
		machine initializedMachine,
		isFirstControlplane bool,
		to common.Software) (func() error, error) {

		id := machine.infra.ID()
		machineLogger := logger.WithFields(map[string]interface{}{
			"machine": id,
		})

		waitForNodeAgent := func() error {
			machineLogger.Info(false, "Waiting for software to be ensured")
			return nil
		}

		ensureJoinSoftware := func() error {
			machine.desiredNodeagent.Software.Merge(to)
			machineLogger.WithFields(map[string]interface{}{
				"current": k8s.Current(machine.currentNodeagent.Software),
				"desired": to,
			}).Info(true, "Join software desired")
			return nil
		}

		ensureKubeadm := func() error {
			machine.desiredNodeagent.Software.Kubeadm = common.Package{
				Version: to.Kubeadm.Version,
			}
			machineLogger.WithFields(map[string]interface{}{
				"current": machine.currentNodeagent.Software.Kubeadm.Version,
				"desired": to.Kubeadm.Version,
			}).Info(true, "Kubeadm desired")
			return nil
		}

		ensureSoftware := func(k8sNode *v1.Node, isControlplane bool, isFirstControlplane bool) func() error {
			return func() (err error) {

				defer func() {
					err = errors.Wrapf(err, "ensuring software on node %s failed", machine.infra.ID())
				}()

				if !isControlplane {
					if err := k8sClient.Drain(k8sNode); err != nil {
						return err
					}
				}

				upgradeAction := "node"
				if isFirstControlplane {
					machineLogger.Info(false, "Migrating kubelet configuration on first controlplane node")
					upgradeAction = fmt.Sprintf("apply %s --yes", to.Kubelet.Version)
					machineLogger.Info(true, "Kubelet configuration on first controlplane node migrated")
				}

				machineLogger.Info(false, "Migrating node")

				_, err = machine.infra.Execute(nil, nil, fmt.Sprintf("sudo kubeadm upgrade %s", upgradeAction))
				if err != nil {
					return err
				}

				machineLogger.WithFields(map[string]interface{}{
					"from": machine.currentNodeagent.Software.Kubelet.Version,
					"to":   to.Kubelet.Version,
				}).Info(true, "Node migrated")

				machine.desiredNodeagent.Software.Merge(to)

				machineLogger.WithFields(map[string]interface{}{
					"from": machine.currentNodeagent.Software.Kubelet.Version,
					"to":   to.Kubelet.Version,
				}).Info(true, "Kubelet desired")
				return nil
			}
		}

		ensureOnline := func(k8sNode *v1.Node) func() error {
			return func() error {
				if err := k8sClient.Uncordon(k8sNode); err != nil {
					return err
				}
				return nil
			}
		}

		if !machine.currentNodeagent.NodeIsReady {
			return waitForNodeAgent, nil
		}

		k8sNode, err := k8sClient.GetNode(id)
		if k8sNode == nil || err != nil {
			if machine.currentNodeagent.Software.Contains(to) {
				return nil, nil
			}
			return ensureJoinSoftware, nil
		}

		k8sNodeIsReady := false
		for _, cond := range k8sNode.Status.Conditions {
			if cond.Type == v1.NodeReady {
				k8sNodeIsReady = true
				break
			}
		}
		if !k8sNodeIsReady {
			// This is a joiners case and treated as up-to-date here
			return nil, nil
		}

		if machine.currentNodeagent.Software.Kubeadm.Version != to.Kubeadm.Version {
			return ensureKubeadm, nil
		}

		isControlplane := machine.tier == Controlplane
		if k8sNode.Status.NodeInfo.KubeletVersion != to.Kubelet.Version {
			return ensureSoftware(k8sNode, isControlplane, isFirstControlplane), nil
		}

		if k8sNode.Spec.Unschedulable && !isControlplane {
			return ensureOnline(k8sNode), nil
		}

		if !machine.currentNodeagent.NodeIsReady || !machine.currentNodeagent.Software.Contains(to) {
			return waitForNodeAgent, nil
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

	logger.WithFields(map[string]interface{}{
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
