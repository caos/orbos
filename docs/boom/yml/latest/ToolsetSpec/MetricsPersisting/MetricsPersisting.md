# MetricsPersisting 
 

## Structure 
 

| Attribute        | Description                                                                         | Default | Collection | Map  |
| ---------------- | ----------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy           | Flag if tool should be deployed                                                     |  false  |            |      |
| metrics          | Spec to define which metrics should get scraped , [here](Metrics/Metrics.md)        |  nil    |            |      |
| storage          | Spec to define how the persistence should be handled , [here](storage/Spec/Spec.md) |  nil    |            |      |
| remoteWrite      | Configuration to write to remote prometheus , [here](RemoteWrite/RemoteWrite.md)    |         |            |      |
| externalLabels   | Static labels added to metrics                                                      |         |            | X    |
| nodeSelector     | NodeSelector for statefulset                                                        |         |            | X    |
| tolerations      | Tolerations to run prometheus on nodes , [here](k8s/Tolerations/Tolerations.md)     |         | X          |      |
| resources        | Resource requirements , [here](k8s/Resources/Resources.md)                          |         |            |      |
| overwriteImage   | Overwrite used image                                                                |         |            |      |
| overwriteVersion | Overwrite used image version                                                        |         |            |      |