package e2e_test

import (
	"fmt"

	"github.com/go-test/deep"
	"gopkg.in/yaml.v3"
)

type readyPods struct {
	orbiter, boom, ambassador, argocd, grafana, kubeStateMetrics,
	fluentbit, fluentd, loggingOperator, loki, prometheusNodeExporter,
	prometheus, prometheusOperator, systemdExporter,
	total int8
}

func assertReadyPods(kubectl kubectlCmd, expectMasters, expectWorkers uint8) (expect readyPods, count func() readyPods) {
	expect = readyPods{
		orbiter:                1,
		boom:                   1,
		ambassador:             2,
		argocd:                 4,
		grafana:                1,
		kubeStateMetrics:       1,
		fluentbit:              int8(expectWorkers),
		fluentd:                1,
		loggingOperator:        1,
		loki:                   1,
		prometheusNodeExporter: int8(expectMasters + expectWorkers),
		prometheus:             1,
		prometheusOperator:     1,
		systemdExporter:        int8(expectMasters + expectWorkers),
		total:                  15 + 3*int8(expectWorkers) + 2*int8(expectMasters),
	}
	return expect, func() readyPods {
		actual := readyPods{
			orbiter:                countReadyPods(kubectl, "app.kubernetes.io/name=orbiter"),
			boom:                   countReadyPods(kubectl, "app.kubernetes.io/name=boom"),
			ambassador:             countReadyPods(kubectl, "app.kubernetes.io/instance=ambassador"),
			argocd:                 countReadyPods(kubectl, "app.kubernetes.io/instance=argocd"),
			grafana:                countReadyPods(kubectl, "app.kubernetes.io/instance=grafana"),
			kubeStateMetrics:       countReadyPods(kubectl, "app.kubernetes.io/instance=kube-state-metrics"),
			fluentbit:              countReadyPods(kubectl, "app.kubernetes.io/name=fluentbit"),
			fluentd:                countReadyPods(kubectl, "app.kubernetes.io/name=fluentd"),
			loggingOperator:        countReadyPods(kubectl, "app.kubernetes.io/instance=logging-operator"),
			loki:                   countReadyPods(kubectl, "app=loki"),
			prometheusNodeExporter: countReadyPods(kubectl, "app=prometheus-node-exporter"),
			prometheus:             countReadyPods(kubectl, "app=prometheus"),
			prometheusOperator:     countReadyPods(kubectl, "app=prometheus-operator-operator"),
			systemdExporter:        countReadyPods(kubectl, "app=systemd-exporter"),
			total:                  countReadyPods(kubectl, ""),
		}

		deep.CompareUnexportedFields = true
		for _, diff := range deep.Equal(actual, expect) {
			fmt.Println("ready pods for", diff)
		}

		return actual
	}
}

func countReadyPods(kubectl kubectlCmd, selector string) (readyPodsCount int8) {

	args := []string{
		"get", "pods",
		"--namespace", "caos-system",
		"--output", "yaml",
	}

	if selector != "" {
		args = append(args, "--selector", selector)
	}

	cmd := kubectl(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return -1
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

	if err := yaml.Unmarshal(out, &pods); err != nil {
		return -1
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		for j := range pod.Status.Conditions {
			condition := pod.Status.Conditions[j]
			if condition.Type != "Ready" {
				continue
			}
			if condition.Status == "True" {
				readyPodsCount++
				break
			}
		}
	}

	return readyPodsCount
}
