# ToolsetSpec 
 

 BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.


## Structure 
 

| Attribute              | Description                                                                                                                              | Default | Collection | Map  |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| boom                   | Boom self reconciling specs , [here](latest/Boom/Boom.md)                                                                                |         |            |      |
| forceApply             | Flag if --force should be used by apply of resources                                                                                     |         |            |      |
| currentStatePath       | Relative folder path where the currentstate is written to                                                                                |         |            |      |
| preApply               | Spec for the yaml-files applied before the applications, for example used secrets , [here](latest/Apply/Apply.md)                        |         |            |      |
| postApply              | Spec for the yaml-files applied after the applications, for example additional crds for the applications , [here](latest/Apply/Apply.md) |         |            |      |
| metricCollection       | Spec for the Prometheus-Operator , [here](latest/MetricCollection/MetricCollection.md)                                                   |         |            |      |
| logCollection          | Spec for the Banzaicloud Logging-Operator , [here](LogCollection/LogCollection.md)                                                       |         |            |      |
| nodeMetricsExporter    | Spec for the Prometheus-Node-Exporter , [here](latest/NodeMetricsExporter/NodeMetricsExporter.md)                                        |         |            |      |
| systemdMetricsExporter | Spec for the Prometheus-Systemd-Exporter , [here](latest/SystemdMetricsExporter/SystemdMetricsExporter.md)                               |         |            |      |
| monitoring             | Spec for the Grafana , [here](monitoring/Monitoring/Monitoring.md)                                                                       |         |            |      |
| apiGateway             | Spec for the Ambassador , [here](latest/APIGateway/APIGateway.md)                                                                        |         |            |      |
| kubeMetricsExporter    | Spec for the Kube-State-Metrics , [here](latest/KubeMetricsExporter/KubeMetricsExporter.md)                                              |         |            |      |
| reconciling            | Spec for the Argo-CD , [here](reconciling/Reconciling/Reconciling.md)                                                                    |         |            |      |
| metricsPersisting      | Spec for the Prometheus instance , [here](latest/MetricsPersisting/MetricsPersisting.md)                                                 |         |            |      |
| logsPersisting         | Spec for the Loki instance , [here](latest/LogsPersisting/LogsPersisting.md)                                                             |         |            |      |
| metricsServer          | Spec for Metrics-Server , [here](latest/MetricsServer/MetricsServer.md)                                                                  |         |            |      |