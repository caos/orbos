# MetricsPersisting 
 

## Structure 
 

| Attribute      | Description                                                                    | Default | Collection  |
| -------------- | ------------------------------------------------------------------------------ | ------- | ----------  |
| deploy         | Flag if tool should be deployed                                                |  false  |             |
| metrics        | Spec to define which metrics should get scraped , [here](Metrics.md)           |  nil    |             |
| storage        | Spec to define how the persistence should be handled , [here](storage/Spec.md) |  nil    |             |
| remoteWrite    | Configuration to write to remote prometheus , [here](RemoteWrite.md)           |         |             |
| externalLabels | Static labels added to metrics                                                 |         |             |
| nodeSelector   | NodeSelector for statefulset                                                   |         |             |