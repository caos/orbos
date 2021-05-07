package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/internal/operator/common"

	"github.com/afiskon/promtail-client/promtail"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"gopkg.in/yaml.v3"
)

func isEnsured(machines int, k8sVersion string) func(promtail.Client, newOrbctlCommandFunc) (bool, error) {
	return func(logger promtail.Client, newOrbctl newOrbctlCommandFunc) (bool, error) {

		orbiter := &struct {
			Clusters struct {
				K8s struct {
					Current kubernetes.CurrentCluster
				}
			}
			Providers struct {
				ProviderUnderTest struct {
					Current struct {
						Ingresses struct {
							Httpsingress infra.Address
							Httpingress  infra.Address
							Kubeapi      infra.Address
						}
					}
				} `yaml:"provider-under-test"`
			}
		}{}
		if err := readYaml(logger, newOrbctl, "caos-internal/orbiter/current.yml", orbiter); err != nil {
			return false, err
		}

		nodeagents := &common.NodeAgentsCurrentKind{}
		if err := readYaml(logger, newOrbctl, "caos-internal/orbiter/node-agents-current.yml", nodeagents); err != nil {
			return false, err
		}

		cluster := orbiter.Clusters.K8s.Current
		currentMachinesLen := len(cluster.Machines.M)

		if cluster.Status != "running" ||
			currentMachinesLen != machines ||
			len(nodeagents.Current.NA) != machines {
			return false, nil
		}

		for _, machine := range cluster.Machines.M {
			if !machine.Joined ||
				!machine.FirewallIsReady ||
				!machine.Ready {
				return false, nil
			}
		}

		for _, nodeagent := range nodeagents.Current.NA {
			if !nodeagent.NodeIsReady ||
				nodeagent.Software.Kubelet.Version != k8sVersion ||
				nodeagent.Software.Kubeadm.Version != k8sVersion ||
				nodeagent.Software.Kubectl.Version != k8sVersion {
				return false, nil
			}
		}

		ep := orbiter.Providers.ProviderUnderTest.Current.Ingresses.Httpsingress

		msg, err := helpers.Check("https", ep.Location, ep.FrontendPort, "/ambassador/v0/check_ready", 200, false)
		if err != nil {
			logger.Warnf("ambassador ready check failed with message: %s: %s", msg, err.Error())
			return false, nil
		}
		logger.Infof(msg)

		return true, nil
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
