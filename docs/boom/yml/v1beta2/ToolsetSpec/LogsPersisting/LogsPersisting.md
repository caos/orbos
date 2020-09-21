# LogsPersisting 
 

## Structure 
 

| Attribute     | Description                                                         | Default | Collection | Map  |
| ------------- | ------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy        | Flag if tool should be deployed                                     |  false  |            |      |
| logs          | Spec to define which logs will get persisted , [here](Logs/Logs.md) |  nil    |            |      |
| clusterOutput | Flag if loki-output should be a clusteroutput instead a output crd  |  false  |            |      |
| nodeSelector  | NodeSelector for statefulset                                        |         |            | X    |