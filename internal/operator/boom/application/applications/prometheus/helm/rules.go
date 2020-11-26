package helm

import "gopkg.in/yaml.v3"

func GetDefaultRules(labels map[string]interface{}) (*AdditionalPrometheusRules, error) {
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
- name: caos.rules
  rules:
   - expr: dist_node_boot_time_seconds
     record: caos_node_boot_time_seconds
   - expr: floor(avg_over_time(dist_systemd_unit_active[5m])+0.2)
     record: caos_systemd_unit_active
   - expr: min(min_over_time(caos_systemd_unit_active[5m])) by (instance)
     record: caos_systemd_ryg
   - expr: avg(max_over_time(caos_probe{type="Upstream",name!="httpingress"}[1m])) by (name)
     record: caos_upstream_probe_ryg
   - expr: max_over_time(caos_probe{type="VIP"}[1m])
     record: caos_vip_probe_ryg
   - expr: sum(1 - avg(rate(dist_node_cpu_seconds_total[5m])))
     record: caos_cluster_cpu_utilisation_5m
   - expr: 100 - (avg by (instance) (irate(dist_node_cpu_seconds_total[5m])) * 100)
     record: caos_node_cpu_utilisation_5m
   - expr: (clamp_max(clamp_min(100-caos_node_cpu_utilisation_5m, 10),20)-10)/10
     record: caos_node_cpu_ryg
   - expr: |-
       sum by (instance) (100 -
       (
         dist_node_memory_MemAvailable_bytes
       /
         dist_node_memory_MemTotal_bytes
       * 100
       ))
     record: caos_node_memory_utilisation
   - expr: (clamp_max(clamp_min(100-caos_node_memory_utilisation, 10),20)-10)/10
     record: caos_node_memory_ryg
   - expr: |-
      100 - (
       min by (instance) (dist_node_filesystem_avail_bytes)
       / min by (instance) (dist_node_filesystem_size_bytes)
       * 100)
     record: caos_node_disk_utilisation
   - expr: dist_kube_node_status_condition
     record: caos_node_ready
   - expr: min_over_time(caos_node_ready[5m])
     record: caos_k8s_node_ryg
   - expr: dist_etcd_server_has_leader or on(instance) up{job="caos_remote_etcd"}
     record: caos_etcd_server_has_leader_and_is_up
   - expr: min_over_time(caos_etcd_server_has_leader_and_is_up[5m])
     record: caos_etcd_ryg
   - expr: |-
       clamp_max(
         clamp_min(
           (
             max_over_time(dist_kube_deployment_status_replicas_available{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_deployment_spec_replicas{namespace=~"(kube-system|caos-system|caos-zitadel)"} or
             max_over_time(dist_kube_statefulset_status_replicas_ready{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_statefulset_replicas{namespace=~"(kube-system|caos-system|caos-zitadel)"} or
             max_over_time(dist_kube_daemonset_status_number_available{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_daemonset_status_desired_number_scheduled{namespace=~"(kube-system|caos-system|caos-zitadel)"}
           ) + 
           1,
           0
         ),
         1
       )
     record: caos_ready_pods_ryg
   - expr: |-
       clamp_max(
         clamp_min(
           (
             max_over_time(dist_kube_deployment_status_replicas{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_deployment_spec_replicas{namespace=~"(kube-system|caos-system|caos-zitadel)"} or
             max_over_time(dist_kube_statefulset_status_replicas_current{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_statefulset_replicas{namespace=~"(kube-system|caos-system|caos-zitadel)"} or
             max_over_time(dist_kube_daemonset_status_current_number_scheduled{namespace=~"(kube-system|caos-system|caos-zitadel)"}[5m]) -
             dist_kube_daemonset_status_desired_number_scheduled{namespace=~"(kube-system|caos-system|caos-zitadel)"}
           ) +
           1,
           0
         ),
         1
       )          
     record: caos_scheduled_pods_ryg
   - expr: |-
       sum(dist_kube_deployment_spec_replicas) + sum(dist_kube_statefulset_replicas) + sum(dist_kube_daemonset_status_desired_number_scheduled) 
     record: caos_desired_pods
   - expr: |-
       sum(dist_kube_deployment_status_replicas) + sum(dist_kube_statefulset_status_replicas_current) + sum(dist_kube_daemonset_status_current_number_scheduled)
     record: caos_scheduled_pods
   - expr: |-
       sum(dist_kube_deployment_status_replicas_available) + sum(dist_kube_statefulset_status_replicas_ready) + sum(dist_kube_daemonset_status_number_available)
     record: caos_ready_pods

# ZITADEL CockroachDB Runtime
   - record: cr_runtime_pod_flapping
     expr: resets(sys_uptime{job="cockroachdb"}[10m])
   - record: cr_runtime_pod_flapping_ryg
     expr: clamp_min(1 - cr_runtime_pod_flapping, 0)
   - record: cr_runtime_high_open_fd_count
     expr: sys_fd_open{job="cockroachdb"} / sys_fd_softlimit{job="cockroachdb"}
   - record: cr_runtime_high_open_fd_count_ryg
     expr: (1 - (clamp_max(clamp_min(floor(cr_runtime_high_open_fd_count * 100), 20), 30) - 20) / 10)
   - record: caos_cr_runtime_pods_ryg
     expr: |-
       cr_runtime_pod_flapping_ryg
       * cr_runtime_high_open_fd_count_ryg

   - record: cr_runtime_version_mismatches
     expr: count by(cluster) (count_values by(tag, cluster) ("version", build_timestamp{job="cockroachdb"}))
   - record: cr_runtime_version_mismatches_ryg
     expr: clamp_min(2 - cr_runtime_version_mismatches, 0)
   - record: caos_cr_runtime_cluster_ryg
     expr: |-
       min(caos_ready_pods_ryg{controller="cockroachdb",namespace="caos-zitadel"})
       * min(caos_scheduled_pods_ryg{controller="cockroachdb",namespace="caos-zitadel"})
       * cr_runtime_version_mismatches

# ZITADEL CockroachDB Capacity
   - record: node:capacity
     expr: sum without(store) (capacity{job="cockroachdb"})
   - record: cluster:capacity
     expr: sum without(instance) (node:capacity{job="cockroachdb"})
   - record: node:capacity_available
     expr: sum without(store) (capacity_available{job="cockroachdb"})
   - record: cluster:capacity_available
     expr: sum without(instance) (node:capacity_available{job="cockroachdb"})
   - record: capacity_available:ratio
     expr: capacity_available{job="cockroachdb"} / capacity{job="cockroachdb"}
   - record: node:capacity_available:ratio
     expr: node:capacity_available{job="cockroachdb"} / node:capacity{job="cockroachdb"}
   - record: cluster:capacity_available:ratio
     expr: cluster:capacity_available{job="cockroachdb"} / cluster:capacity{job="cockroachdb"}

   - record: caos_cr_capacity_store_ryg
     expr: (clamp_max(clamp_min(100 * capacity_available:ratio{job="cockroachdb"}, 15), 30)-15) /15
   - record: caos_cr_capacity_cluster_ryg
     expr: (clamp_max(clamp_min(100 * cluster:capacity_available:ratio{job="cockroachdb"}, 20), 30) - 20) /10

# ZITADEL CockroachDB Replicas
   - record: cr_replicas_clock_offset_near_max_ryg
     expr: 1 - (clamp_max(clamp_min(abs(clock_offset_meannanos), 200 * 1000 * 1000), 300 * 1000 * 1000) - 200 * 1000 * 1000) / (100 * 1000 * 1000)
   - record: cr_ca_certificate_expires_soon_ryg
     expr: (clamp_max(clamp_min(security_certificate_expiration_ca{job="cockroachdb"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
#   - record: cr_client_ca_certificate_expires_soon_ryg
#     expr: (clamp_max(clamp_min(security_certificate_expiration_client_ca{job="cockroachdb"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
   - record: cr_node_certificate_expires_soon_ryg
     expr: (clamp_max(clamp_min(security_certificate_expiration_node{job="cockroachdb"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
#   - record: cr_node_client_certificate_expires_soon_ryg
#     expr: (clamp_max(clamp_min(cr_node_client_certificate_expires_soon{job="cockroachdb"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
   - record: caos_cr_replicas_nodes_ryg
     expr: |-
       cr_replicas_clock_offset_near_max_ryg
       * cr_ca_certificate_expires_soon_ryg
       * cr_node_certificate_expires_soon_ryg
#       * cr_client_ca_certificate_expires_soon_ryg
#       * cr_node_client_certificate_expires_soon_ryg
   - record: caos_cr_replicas_stores_ryg
     expr: |-
       clamp_min(1 - requests_slow_latch{job="cockroachdb"}, 0)
       * clamp_min(1 - requests_slow_lease{job="cockroachdb"}, 0)
       * clamp_min(1 - requests_slow_raft{job="cockroachdb"}, 0)

# Cluster RYG
   - record: caos_orb_ryg
     expr: |-
       # Platform Monitoring
       min(caos_node_cpu_ryg)
       * min(caos_systemd_ryg)
       * min(caos_vip_probe_ryg)
       * min(caos_upstream_probe_ryg)
       * min(caos_node_memory_ryg)
       * min(caos_k8s_node_ryg)
       * avg(caos_etcd_ryg)
       * min(caos_ready_pods_ryg{namespace=~"(kube-system|caos-system|caos-zitadel)"})
       * min(caos_scheduled_pods_ryg{namespace=~"(kube-system|caos-system|caos-zitadel)"})
       # ZITADEL Monitoring
       * min(caos_cr_runtime_cluster_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0)) 
       * min(caos_cr_runtime_pods_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0))
       * min(caos_cr_replicas_nodes_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0))
       * min(caos_cr_replicas_stores_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0))
       * min(caos_cr_capacity_cluster_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0))
       * min(caos_cr_capacity_store_ryg) or on() 1 - (up{endpoint="http",instance="orbos-n6st1n",job="cockroachdb",namespace="caos-zitadel",pod="cockroachdb-0",service="cockroachdb"} or on() vector(0))
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
