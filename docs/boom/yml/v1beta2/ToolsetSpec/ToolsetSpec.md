# ToolsetSpec 
 

 BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.


## Structure 
 

| Attribute              | Description                                                                                               | Default | Collection | Map  |
| ---------------------- | --------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| boom                   | Boom self reconciling specs                                                                               |         |            |      |
| forceApply             | Flag if --force should be used by apply of resources                                                      |         |            |      |
| currentStatePath       | Relative folder path where the currentstate is written to                                                 |         |            |      |
| preApply               | Spec for the yaml-files applied before the applications, for example used secrets                         |         |            |      |
| postApply              | Spec for the yaml-files applied after the applications, for example additional crds for the applications  |         |            |      |
| metricCollection       | Spec for the Prometheus-Operator                                                                          |         |            |      |
| logCollection          | Spec for the Banzaicloud Logging-Operator , [here](LogCollection/LogCollection.md)                        |         |            |      |
| nodeMetricsExporter    | Spec for the Prometheus-Node-Exporter                                                                     |         |            |      |
| systemdMetricsExporter | Spec for the Prometheus-Systemd-Exporter                                                                  |         |            |      |
| monitoring             | Spec for the Grafana                                                                                      |         |            |      |
| apiGateway             | Spec for the Ambassador                                                                                   |         |            |      |
| kubeMetricsExporter    | Spec for the Kube-State-Metrics                                                                           |         |            |      |
| reconciling            | Spec for the Argo-CD                                                                                      |         |            |      |
| metricsPersisting      | Spec for the Prometheus instance                                                                          |         |            |      |
| logsPersisting         | Spec for the Loki instance                                                                                |         |            |      |
| metricsServer          | Spec for Metrics-Server                                                                                   |         |            |      |