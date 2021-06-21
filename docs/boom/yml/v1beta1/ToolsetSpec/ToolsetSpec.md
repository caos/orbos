# ToolsetSpec 
 

 BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.


## Structure 
 

| Attribute                   | Description                                                                                                                       | Default | Collection | Map  |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| boomVersion                 | Version of BOOM which should be reconciled                                                                                        |         |            |      |
| forceApply                  | Flag if --force should be used by apply of resources                                                                              |         |            |      |
| currentStatePath            | Relative folder path where the currentstate is written to                                                                         |         |            |      |
| preApply                    | Spec for the yaml-files applied before the applications, for example used secrets , [here](Apply/Apply.md)                        |         |            |      |
| postApply                   | Spec for the yaml-files applied after the applications, for example additional crds for the applications , [here](Apply/Apply.md) |         |            |      |
| prometheus-operator         | Spec for the Prometheus-Operator , [here](PrometheusOperator/PrometheusOperator.md)                                               |         |            |      |
| logging-operator            | Spec for the Banzaicloud Logging-Operator , [here](LoggingOperator/LoggingOperator.md)                                            |         |            |      |
| prometheus-node-exporter    | Spec for the Prometheus-Node-Exporter , [here](PrometheusNodeExporter/PrometheusNodeExporter.md)                                  |         |            |      |
| prometheus-systemd-exporter | Spec for the Prometheus-Systemd-Exporter , [here](PrometheusSystemdExporter/PrometheusSystemdExporter.md)                         |         |            |      |
| grafana                     | Spec for the Grafana                                                                                                              |         |            |      |
| ambassador                  | Spec for the Ambassador , [here](Ambassador/Ambassador.md)                                                                        |         |            |      |
| kube-state-metrics          | Spec for the Kube-State-Metrics , [here](KubeStateMetrics/KubeStateMetrics.md)                                                    |         |            |      |
| argocd                      | Spec for the Argo-CD                                                                                                              |         |            |      |
| prometheus                  | Spec for the Prometheus instance , [here](Prometheus/Prometheus.md)                                                               |         |            |      |
| loki                        | Spec for the Loki instance , [here](Loki/Loki.md)                                                                                 |         |            |      |
| metrics-server              | Spec for Metrics-Server , [here](MetricsServer/MetricsServer.md)                                                                  |         |            |      |