package kubernetes

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

func ensureK8sVersion(
	logger logging.Logger,
	orbiterCommit string,
	repoURL string,
	repoKey string,
	target k8s.KubernetesVersion,
	k8sClient *k8s.Client,
	currentComputes map[string]*Compute,
	nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent,
	nodeAgentsDesired map[string]*orbiter.NodeAgentSpec,
	controlplane infra.Computes,
	workers infra.Computes) (bool, error) {

	findPath := func(computes infra.Computes) (orbiter.Software, orbiter.Software, error) {

		var overallLowKubelet k8s.KubernetesVersion
		var overallLowKubeletMinor int
		zeroSW := orbiter.Software{}

		for _, cp := range computes {
			node, err := k8sClient.GetNode(cp.ID())
			if err != nil {
				continue
			}

			nodeinfoKubelet := node.Status.NodeInfo.KubeletVersion

			logger.WithFields(map[string]interface{}{
				"compute": cp.ID(),
				"kubelet": nodeinfoKubelet,
			}).Debug("Found kubelet version from node info")
			kubelet := k8s.ParseString(nodeinfoKubelet)
			if kubelet == k8s.Unknown {
				return zeroSW, zeroSW, errors.Errorf("parsing version %s from nodes %s info failed", nodeinfoKubelet, cp.ID())
			}

			kubeletMinor, err := kubelet.ExtractMinor()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting minor from kubelet version %s from nodes %s info failed", nodeinfoKubelet, cp.ID())
			}

			if overallLowKubelet == k8s.Unknown {
				overallLowKubelet = kubelet
				overallLowKubeletMinor = kubeletMinor
				continue
			}

			kubeletPatch, err := kubelet.ExtractPatch()
			if err != nil {
				return zeroSW, zeroSW, errors.Wrapf(err, "extracting patch from kubelet version %s from nodes %s info failed", nodeinfoKubelet, cp.ID())
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
		compute infra.Compute,
		isFirstControlplane bool,
		to orbiter.Software) (func() error, error) {

		desNa, ok := nodeAgentsDesired[compute.ID()]
		if !ok {
			desNa = &orbiter.NodeAgentSpec{}
		}

		if desNa.Software == nil {
			desNa.Software = &orbiter.Software{}
		}

		desiredSoftware := desNa.Software

		currentSoftware := orbiter.Software{}
		naCurrent, ok := nodeAgentsCurrent[compute.ID()]
		if ok {
			currentSoftware = naCurrent.Software
		}

		ensureNodeagent := func(from string) func() error {
			return func() error {
				logger.WithFields(map[string]interface{}{
					"compute": compute.ID(),
					"from":    from,
					"to":      orbiterCommit,
				}).Info("Ensuring node agent")

				return errors.Wrap(installNodeAgent(logger, compute, repoURL, repoKey), "upgrading node agent failed")
			}
		}

		ensureKubeadm := func() error {
			desiredSoftware.Kubeadm = orbiter.Package{
				Version: to.Kubeadm.Version,
			}
			logger.WithFields(map[string]interface{}{
				"compute": compute.ID(),
				"from":    currentSoftware.Kubeadm.Version,
				"to":      to.Kubeadm.Version,
			}).Info("Ensuring kubeadm")
			return nil
		}

		ensureSoftware := func(k8sNode *v1.Node, isControlplane bool, isFirstControlplane bool) func() error {
			return func() (err error) {

				defer func() {
					err = errors.Wrapf(err, "ensuring software on node %s failed", compute.ID())
				}()

				id := compute.ID()
				if !isControlplane {
					logger.WithFields(map[string]interface{}{
						"compute": id,
					}).Info("Draining node")

					if err := k8sClient.Drain(k8sNode); err != nil {
						return err
					}
				}

				upgradeAction := "node"
				if isFirstControlplane {
					logger.WithFields(map[string]interface{}{
						"compute": id,
					}).Info("Upgrading kubelet configuration on first controlplane node")

					upgradeAction = fmt.Sprintf("apply %s --yes", to.Kubelet.Version)
				} else {
					logger.WithFields(map[string]interface{}{
						"compute": id,
					}).Info("Migrating node")
				}

				_, err = compute.Execute(nil, nil, fmt.Sprintf("sudo kubeadm upgrade %s", upgradeAction))
				if err != nil {
					return err
				}

				logger.WithFields(map[string]interface{}{
					"compute": id,
					"from":    currentSoftware.Kubelet.Version,
					"to":      to.Kubelet.Version,
				}).Info("Ensuring kubelet")

				desiredSoftware.Merge(to)
				return nil
			}
		}

		ensureOnline := func(k8sNode *v1.Node) func() error {
			return func() error {
				logger.WithFields(map[string]interface{}{
					"compute": compute.ID(),
				}).Info("Bringing node back online")
				return k8sClient.Uncordon(k8sNode)
			}
		}

		id := compute.ID()

		var response []byte
		isActive := "sudo systemctl is-active node-agentd"
		err := try(logger, time.NewTimer(7*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
			var cbErr error
			response, cbErr = cmp.Execute(nil, nil, isActive)
			return errors.Wrapf(cbErr, "remote command %s returned an unsuccessful exit code", isActive)
		})
		logger.WithFields(map[string]interface{}{
			"command":  isActive,
			"response": string(response),
		}).Debug("Executed command")
		if err != nil && !strings.Contains(string(response), "activating") {
			return ensureNodeagent("not running"), nil
		}

		if naCurrent.Commit != orbiterCommit {
			showVersion := "node-agent --version"

			err := try(logger, time.NewTimer(7*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
				var cbErr error
				response, cbErr = cmp.Execute(nil, nil, showVersion)
				return errors.Wrapf(cbErr, "running command %s remotely failed", showVersion)
			})
			logger.WithFields(map[string]interface{}{
				"command":  showVersion,
				"response": string(response),
			}).Debug("Executed command")

			fields := strings.Fields(string(response))
			if err != nil || len(fields) != 1 || fields[0] != orbiterCommit {
				return ensureNodeagent(fields[0]), nil
			}
		}

		k8sNode, err := k8sClient.GetNode(id)
		if k8sNode == nil || err != nil {
			// This is a joiners case and treated as up-to-date here
			return nil, nil
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

		if currentSoftware.Kubeadm.Version != to.Kubeadm.Version {
			return ensureKubeadm, nil
		}

		isControlplane := currentComputes[id].Metadata.Tier == Controlplane
		if k8sNode.Status.NodeInfo.KubeletVersion != to.Kubelet.Version {
			return ensureSoftware(k8sNode, isControlplane, isFirstControlplane), nil
		}

		if k8sNode.Spec.Unschedulable && !isControlplane {
			return ensureOnline(k8sNode), nil
		}
		return nil, nil
	}

	sort.Sort(controlplane)
	sort.Sort(workers)

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
	for idx, compute := range append(controlplane, workers...) {

		next, err := plan(compute, idx == 0, to)
		if err != nil {
			return false, errors.Wrapf(err, "planning compute %s failed", compute.ID())
		}

		if next == nil || !nexting {
			currentComputes[compute.ID()].Status = "running"
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
