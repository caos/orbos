# Loki 
 

## Structure 
 

| Attribute     | Description                                                         | Default | Collection | Map  |
| ------------- | ------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy        | Flag if tool should be deployed                                     |  false  |            |      |
| logs          | Spec to define which logs will get persisted , [here](Logs/Logs.md) |  nil    |            |      |
| storage       | Spec to define how the persistence should be handled                |  nil    |            |      |
| clusterOutput | Flag if loki-output should be a clusteroutput instead a output crd  |  false  |            |      |
| nodeSelector  | NodeSelector for statefulset                                        |         |            | X    |
| tolerations   | Tolerations to run loki on nodes                                    |         | X          |      |