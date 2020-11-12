# Metrics 
 

 When the metrics spec is nil all metrics will get scraped.


## Structure 
 

| Attribute                   | Description                                                          | Default | Collection | Map  |
| --------------------------- | -------------------------------------------------------------------- | ------- | ---------- | ---  |
| ambassador                  | Bool if metrics should get scraped from ambassador                   |         |            |      |
| argocd                      | Bool if metrics should get scraped from argo-cd                      |         |            |      |
| kube-state-metrics          | Bool if metrics should get scraped from kube-state-metrics           |         |            |      |
| prometheus-node-exporter    | Bool if metrics should get scraped from prometheus-node-exporter     |         |            |      |
| prometheus-systemd-exporter | Bool if metrics should get scraped from prometheus-systemd-exporter  |         |            |      |
| api-server                  | Bool if metrics should get scraped from kube-api-server              |         |            |      |
| prometheus-operator         | Bool if metrics should get scraped from prometheus-operator          |         |            |      |
| logging-operator            | Bool if metrics should get scraped from logging-operator             |         |            |      |
| loki                        | Bool if metrics should get scraped from loki                         |         |            |      |
| boom                        | Bool if metrics should get scraped from boom                         |         |            |      |
| orbiter                     | Bool if metrics should get scraped from orbiter                      |         |            |      |