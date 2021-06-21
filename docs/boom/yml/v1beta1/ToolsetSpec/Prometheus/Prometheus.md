# Prometheus 
 

## Structure 
 

| Attribute    | Description                                                                      | Default | Collection | Map  |
| ------------ | -------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy       | Flag if tool should be deployed                                                  |  false  |            |      |
| metrics      | Spec to define which metrics should get scraped , [here](Metrics/Metrics.md)     |  nil    |            |      |
| storage      | Spec to define how the persistence should be handled                             |  nil    |            |      |
| remoteWrite  | Configuration to write to remote prometheus , [here](RemoteWrite/RemoteWrite.md) |         |            |      |
| nodeSelector | NodeSelector for statefulset                                                     |         |            | X    |
| tolerations  | Tolerations to run prometheus on nodes                                           |         | X          |      |