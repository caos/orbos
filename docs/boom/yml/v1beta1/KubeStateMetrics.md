# KubeStateMetrics 
 

## Structure 
 

| Attribute    | Description                                                                                | Default | Collection  |
| ------------ | ------------------------------------------------------------------------------------------ | ------- | ----------  |
| deploy       | Flag if tool should be deployed                                                            |  false  |             |
| replicaCount | Number of replicas used for deployment                                                     |  1      |             |
| nodeSelector | NodeSelector for deployment                                                                |         |             |
| tolerations  | Tolerations to run kube state metrics exporter on nodes , [here](toleration/Toleration.md) |         | X           |