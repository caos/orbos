# MetricCollection 
 

## Structure 
 

| Attribute        | Description                                                                              | Default | Collection | Map  |
| ---------------- | ---------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy           | Flag if tool should be deployed                                                          |  false  |            |      |
| nodeSelector     | NodeSelector for deployment                                                              |         |            | X    |
| tolerations      | Tolerations to run prometheus-operator on nodes , [here](k8s/Tolerations/Tolerations.md) |         | X          |      |
| resources        | Resource requirements , [here](k8s/Resources/Resources.md)                               |         |            |      |
| overwriteImage   | Overwrite used image                                                                     |         |            |      |
| overwriteVersion | Overwrite used image version                                                             |         |            |      |