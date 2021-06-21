# ToolsetSpec 
 

 BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.


## Structure 
 

| Attribute              | Description                                                                                                                       | Default | Collection | Map  |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| boom                   | Boom self reconciling specs , [here](Boom/Boom.md)                                                                                |         |            |      |
| forceApply             | Relative folder path where the currentstate is written to                                                                         |         |            |      |
| currentStatePath       | Flag if --force should be used by apply of resources                                                                              |         |            |      |
| preApply               | Spec for the yaml-files applied before the applications, for example used secrets , [here](Apply/Apply.md)                        |         |            |      |
| postApply              | Spec for the yaml-files applied after the applications, for example additional crds for the applications , [here](Apply/Apply.md) |         |            |      |
| metricCollection       | Spec for the Prometheus-Operator , [here](MetricCollection/MetricCollection.md)                                                   |         |            |      |
| logCollection          | Spec for the Banzaicloud Logging-Operator , [here](LogCollection/LogCollection.md)                                                |         |            |      |
| nodeMetricsExporter    | Spec for the Prometheus-Node-Exporter , [here](NodeMetricsExporter/NodeMetricsExporter.md)                                        |         |            |      |
| systemdMetricsExporter | Spec for the Prometheus-Systemd-Exporter , [here](SystemdMetricsExporter/SystemdMetricsExporter.md)                               |         |            |      |
| apiGateway             | Spec for the Ambassador , [here](APIGateway/APIGateway.md)                                                                        |         |            |      |
| kubeMetricsExporter    | Spec for the Kube-State-Metrics , [here](KubeMetricsExporter/KubeMetricsExporter.md)                                              |         |            |      |
| metricsPersisting      | Spec for the Prometheus instance , [here](MetricsPersisting/MetricsPersisting.md)                                                 |         |            |      |
| logsPersisting         | Spec for the Loki instance , [here](LogsPersisting/LogsPersisting.md)                                                             |         |            |      |