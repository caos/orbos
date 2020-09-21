# MetricsPersisting 
 

## Structure 
 

| Attribute      | Description                                                                      | Default | Collection | Map  |
| -------------- | -------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy         | Flag if tool should be deployed                                                  |  false  |            |      |
| metrics        | Spec to define which metrics should get scraped , [here](Metrics/Metrics.md)     |  nil    |            |      |
| remoteWrite    | Configuration to write to remote prometheus , [here](RemoteWrite/RemoteWrite.md) |         |            |      |
| externalLabels | Static labels added to metrics                                                   |         |            | X    |
| nodeSelector   | NodeSelector for statefulset                                                     |         |            | X    |