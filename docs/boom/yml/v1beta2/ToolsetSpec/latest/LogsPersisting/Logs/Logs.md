# Logs 
 

 When the logs spec is nil all logs will get persisted in loki.


## Structure 
 

| Attribute                   | Description                                                      | Default | Collection | Map  |
| --------------------------- | ---------------------------------------------------------------- | ------- | ---------- | ---  |
| ambassador                  | Bool if logs will get persisted for ambassador                   |         |            |      |
| grafana                     | Bool if logs will get persisted for grafana                      |         |            |      |
| argocd                      | Bool if logs will get persisted for argo-cd                      |         |            |      |
| kube-state-metrics          | Bool if logs will get persisted for kube-state-metrics           |         |            |      |
| prometheus-node-exporter    | Bool if logs will get persisted for prometheus-node-exporter     |         |            |      |
| prometheus-operator         | Bool if logs will get persisted for prometheus-operator          |         |            |      |
| prometheus-systemd-exporter | Bool if logs will get persisted for Prometheus-Systemd-Exporter  |         |            |      |
| logging-operator            | Bool if logs will get persisted for logging-operator             |         |            |      |
| loki                        | Bool if logs will get persisted for loki                         |         |            |      |
| prometheus                  | Bool if logs will get persisted for prometheus                   |         |            |      |
| metrics-server              | Bool if logs will get persisted for the metrics-secret           |         |            |      |