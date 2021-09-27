package metrics

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/nodemetricsexporter/info"
	"github.com/caos/orbos/v5/internal/operator/boom/labels"
)

func getLocalServiceMonitor(appName string, monitorMatchingLabels, serviceMatchingLabels map[string]string) *servicemonitor.Config {
	relabelings := make([]*servicemonitor.ConfigRelabeling, 0)
	relabeling := &servicemonitor.ConfigRelabeling{
		Action:       "replace",
		Regex:        "(.*)",
		Replacement:  "$1",
		SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
		TargetLabel:  "instance",
	}
	relabelings = append(relabelings, relabeling)

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:        "metrics",
		Path:        "/metrics",
		Relabelings: relabelings,
	}

	return &servicemonitor.Config{
		Name:                  "prometheus-local-node-exporter-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		JobName:               fmt.Sprintf("local-%s", appName),
		MonitorMatchingLabels: monitorMatchingLabels,
		ServiceMatchingLabels: serviceMatchingLabels,
	}
}
func getIngestionServiceMonitor(appName string, monitorMatchingLabels, serviceMatchingLabels map[string]string) *servicemonitor.Config {

	relabelings := []*servicemonitor.ConfigRelabeling{{
		Action:       "replace",
		SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
		TargetLabel:  "instance",
	}, {
		Action:       "replace",
		SourceLabels: []string{"job"},
		TargetLabel:  "job",
		Replacement:  "caos_remote_${1}",
	}, {
		Action: "labeldrop",
		Regex:  "(container|endpoint|namespace|pod)",
	}}

	metricRelabelings := []*servicemonitor.ConfigRelabeling{{
		Action: "keep",
		SourceLabels: []string{
			"__name__",
			"mode",
			"device",
			"fstype",
			"mountpoint",
		},
		Regex: "(node_cpu_seconds_total;idle;;;|(node_filesystem_avail_bytes|node_filesystem_size_bytes);;rootfs;rootfs;/|(node_memory_MemAvailable_bytes|node_memory_MemTotal_bytes|node_boot_time_seconds);;;;)",
	}, {
		Action: "labelkeep",
		Regex:  "__.+|job|instance|cpu",
	}, {
		Action:       "replace",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Replacement:  "dist_${1}",
	}}

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:              "metrics",
		Path:              "/metrics",
		HonorLabels:       false,
		Relabelings:       relabelings,
		MetricRelabelings: metricRelabelings,
	}

	return &servicemonitor.Config{
		Name:                  "prometheus-ingestion-node-exporter-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		JobName:               fmt.Sprintf("ingestion-%s", appName),
		MonitorMatchingLabels: monitorMatchingLabels,
		ServiceMatchingLabels: serviceMatchingLabels,
	}
}

func GetServicemonitors(instanceName string) []*servicemonitor.Config {

	appName := info.GetName()
	appNameStr := appName.String()
	monitorLabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	return []*servicemonitor.Config{
		getLocalServiceMonitor(appNameStr, monitorLabels, ls),
		getIngestionServiceMonitor(appNameStr, monitorLabels, ls),
	}
}
