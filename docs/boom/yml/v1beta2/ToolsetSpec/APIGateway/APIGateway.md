# APIGateway 
 

## Structure 
 

| Attribute         | Description                                                                        | Default | Collection | Map  |
| ----------------- | ---------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy            | Flag if tool should be deployed                                                    |  false  |            |      |
| replicaCount      | Number of replicas used for deployment                                             |  1      |            |      |
| service           | Service definition for ambassador , [here](AmbassadorService/AmbassadorService.md) |         |            |      |
| activateDevPortal | Activate the dev portal mapping                                                    |         |            |      |
| nodeSelector      | NodeSelector for deployment                                                        |         |            | X    |
| caching           | Caching options , [here](Caching/Caching.md)                                       |         |            |      |