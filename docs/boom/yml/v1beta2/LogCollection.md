# LogCollection 
 

## Structure 
 

| Attribute      | Description                                                                    | Default | Collection  |
| -------------- | ------------------------------------------------------------------------------ | ------- | ----------  |
| deploy         | Flag if tool should be deployed                                                |  false  |             |
| fluentdStorage | Spec to define how the persistence should be handled , [here](storage/Spec.md) |         |             |
| nodeSelector   | NodeSelector for deployment                                                    |         |             |
| resources      | Resource requirements , [here](v1/ResourceRequirements.md)                     |         |             |