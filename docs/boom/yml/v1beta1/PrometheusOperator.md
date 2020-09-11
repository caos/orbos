# PrometheusOperator 
 

## Structure 
 

| Attribute    | Description                                                                        | Default | Collection  |
| ------------ | ---------------------------------------------------------------------------------- | ------- | ----------  |
| deploy       | Flag if tool should be deployed                                                    |  false  |             |
| nodeSelector | NodeSelector for deployment                                                        |         |             |
| tolerations  | Tolerations to run prometheus-operator on nodes , [here](toleration/Toleration.md) |         | X           |