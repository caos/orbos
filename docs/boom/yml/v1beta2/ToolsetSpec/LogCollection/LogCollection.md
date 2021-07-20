# LogCollection 
 

## Structure 
 

| Attribute      | Description                                                                                                             | Default | Collection | Map  |
| -------------- | ----------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy         | Flag if tool should be deployed                                                                                         |  false  |            |      |
| fluentdStorage | Spec to define how the persistence should be handled @deprecated Use Fluentd.PVC instead , [here](storage/Spec/Spec.md) |         |            |      |
| nodeSelector   | NodeSelector for deployment                                                                                             |         |            | X    |
| tolerations    | Tolerations to run fluentbit on nodes , [here](k8s/Tolerations/Tolerations.md)                                          |         | X          |      |
| resources      | Resource requirements , [here](k8s/Resources/Resources.md)                                                              |         | X          |      |