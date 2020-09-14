# LogsPersisting 
 

## Structure 
 

| Attribute     | Description                                                                    | Default | Collection  |
| ------------- | ------------------------------------------------------------------------------ | ------- | ----------  |
| deploy        | Flag if tool should be deployed                                                |  false  |             |
| logs          | Spec to define which logs will get persisted , [here](Logs.md)                 |  nil    |             |
| storage       | Spec to define how the persistence should be handled , [here](storage/Spec.md) |  nil    |             |
| clusterOutput | Flag if loki-output should be a clusteroutput instead a output crd             |  false  |             |
| nodeSelector  | NodeSelector for statefulset                                                   |         |             |