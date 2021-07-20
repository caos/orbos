# LogCollection 
 

## Structure 
 

| Attribute        | Description                                                                                | Default | Collection | Map  |
| ---------------- | ------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| deploy           | Flag if tool should be deployed                                                            |  false  |            |      |
| fluentd          | Fluentd Specs , [here](Fluentd/Fluentd.md)                                                 |         |            |      |
| nodeSelector     | NodeSelector for deployment                                                                |         |            | X    |
| tolerations      | Tolerations to run fluentbit on nodes , [here](k8s/Tolerations/Tolerations.md)             |         | X          |      |
| resources        | Resource requirements , [here](k8s/Resources/Resources.md)                                 |         | X          |      |
| nodeSelector     | NodeSelector for deployment                                                                |         |            | X    |
| tolerations      | Tolerations to run fluentbit on nodes , [here](k8s/Tolerations/Tolerations.md)             |         | X          |      |
| resources        | Resource requirements , [here](k8s/Resources/Resources.md)                                 |         | X          |      |
| clusterOutputs   | ClusterOutputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified  |         | X          |      |
| outputs          | Outputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified         |         | X          |      |
| watchNamespaces  | Watch these namespaces                                                                     |         | X          |      |
| overwriteImage   | Overwrite used image                                                                       |         |            |      |
| overwriteVersion | Overwrite used image version                                                               |         |            |      |