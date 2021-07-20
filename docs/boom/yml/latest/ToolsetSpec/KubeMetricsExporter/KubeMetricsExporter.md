# KubeMetricsExporter 
 

## Structure 
 

| Attribute        | Description                                                                                      | Default | Collection | Map  |
| ---------------- | ------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| deploy           | Flag if tool should be deployed                                                                  |  false  |            |      |
| replicaCount     | Number of replicas used for deployment                                                           |  1      |            |      |
| affinity         | Pod scheduling constrains , [here](k8s/Affinity/Affinity.md)                                     |         | X          |      |
| nodeSelector     | NodeSelector for deployment                                                                      |         |            | X    |
| tolerations      | Tolerations to run kube state metrics exporter on nodes , [here](k8s/Tolerations/Tolerations.md) |         | X          |      |
| resources        | Resource requirements , [here](k8s/Resources/Resources.md)                                       |         | X          |      |
| overwriteImage   | Overwrite used image                                                                             |         |            |      |
| overwriteVersion | Overwrite used image version                                                                     |         |            |      |