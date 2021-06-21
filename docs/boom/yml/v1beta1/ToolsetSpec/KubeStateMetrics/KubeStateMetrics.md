# KubeStateMetrics 
 

## Structure 
 

| Attribute    | Description                                              | Default | Collection | Map  |
| ------------ | -------------------------------------------------------- | ------- | ---------- | ---  |
| deploy       | Flag if tool should be deployed                          |  false  |            |      |
| replicaCount | Number of replicas used for deployment                   |  1      |            |      |
| nodeSelector | NodeSelector for deployment                              |         |            | X    |
| tolerations  | Tolerations to run kube state metrics exporter on nodes  |         | X          |      |