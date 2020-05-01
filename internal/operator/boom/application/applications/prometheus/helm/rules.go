package helm

import "gopkg.in/yaml.v3"

func GetDefaultRules(labels map[string]string) (*AdditionalPrometheusRules, error) {
	rulesStr := `name: node-exporter.rules
groups:
- name: node-exporter.rules
  rules:
  - expr: |-
      count without (cpu) (
        count without (mode) (
          node_cpu_seconds_total{job="node-exporter"}
        )
      )
    record: instance:node_num_cpu:sum
  - expr: |-
      1 - avg without (cpu, mode) (
        rate(node_cpu_seconds_total{job="node-exporter", mode="idle"}[1m])
      )
    record: instance:node_cpu_utilisation:rate1m
  - expr: |-
      (
        node_load1{job="node-exporter"}
      /
        instance:node_num_cpu:sum{job="node-exporter"}
      )
    record: instance:node_load1_per_cpu:ratio
  - expr: |-
      1 - (
        node_memory_MemAvailable_bytes{job="node-exporter"}
      /
        node_memory_MemTotal_bytes{job="node-exporter"}
      )
    record: instance:node_memory_utilisation:ratio
  - expr: rate(node_vmstat_pgmajfault{job="node-exporter"}[1m])
    record: instance:node_vmstat_pgmajfault:rate1m
  - expr: rate(node_disk_io_time_seconds_total{job="node-exporter", device=~"nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+"}[1m])
    record: instance_device:node_disk_io_time_seconds:rate1m
  - expr: rate(node_disk_io_time_weighted_seconds_total{job="node-exporter", device=~"nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+"}[1m])
    record: instance_device:node_disk_io_time_weighted_seconds:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_receive_bytes_total{job="node-exporter", device!="lo"}[1m])
      )
    record: instance:node_network_receive_bytes_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_transmit_bytes_total{job="node-exporter", device!="lo"}[1m])
      )
    record: instance:node_network_transmit_bytes_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_receive_drop_total{job="node-exporter", device!="lo"}[1m])
      )
    record: instance:node_network_receive_drop_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_transmit_drop_total{job="node-exporter", device!="lo"}[1m])
      )
    record: instance:node_network_transmit_drop_excluding_lo:rate1m
- name: node.rules
  rules:
  - expr: sum(min(kube_pod_info) by (node))
    record: ':kube_pod_info_node_count:'
  - expr: max(label_replace(kube_pod_info{job="kube-state-metrics"}, "pod", "$1", "pod", "(.*)")) by (node, namespace, pod)
    record: 'node_namespace_pod:kube_pod_info:'
  - expr: |-
      count by (node) (sum by (node, cpu) (
        node_cpu_seconds_total{job="node-exporter"}
      * on (namespace, pod) group_left(node)
        node_namespace_pod:kube_pod_info:
      ))
    record: node:node_num_cpu:sum
  - expr: |-
      sum(
        node_memory_MemAvailable_bytes{job="node-exporter"} or
        (
          node_memory_Buffers_bytes{job="node-exporter"} +
          node_memory_Cached_bytes{job="node-exporter"} +
          node_memory_MemFree_bytes{job="node-exporter"} +
          node_memory_Slab_bytes{job="node-exporter"}
        )
      )
    record: :node_memory_MemAvailable_bytes:sum
  `

	struc := &AdditionalPrometheusRules{
		AdditionalLabels: labels,
	}
	rulesByte := []byte(rulesStr)
	if err := yaml.Unmarshal(rulesByte, struc); err != nil {
		return nil, err
	}
	return struc, nil
}
