# LoggingOperator 
 

## Structure 
 

| Attribute      | Description                                                                         | Default | Collection | Map  |
| -------------- | ----------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy         | Flag if tool should be deployed                                                     |  false  |            |      |
| fluentdStorage | Spec to define how the persistence should be handled , [here](storage/Spec/Spec.md) |         |            |      |
| nodeSelector   | NodeSelector for deployment                                                         |         |            | X    |
| tolerations    | Tolerations to run fluentbit on nodes , [here](toleration/Toleration/Toleration.md) |         | X          |      |