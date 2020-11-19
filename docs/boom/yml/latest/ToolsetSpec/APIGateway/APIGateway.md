# APIGateway 
 

## Structure 
 

| Attribute         | Description                                                                        | Default | Collection | Map  |
| ----------------- | ---------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy            | Flag if tool should be deployed                                                    |  false  |            |      |
| replicaCount      | Number of replicas used for deployment                                             |  1      |            |      |
| affinity          | Pod scheduling constrains , [here](k8s/Affinity/Affinity.md)                       |         |            |      |
| service           | Service definition for ambassador , [here](AmbassadorService/AmbassadorService.md) |         |            |      |
| activateDevPortal | Activate the dev portal mapping                                                    |         |            |      |
| nodeSelector      | NodeSelector for deployment                                                        |         |            | X    |
| tolerations       | Tolerations to run ambassador on nodes , [here](k8s/Tolerations/Tolerations.md)    |         | X          |      |
| resources         | Resource requirements , [here](k8s/Resources/Resources.md)                         |         |            |      |
| caching           | Caching options , [here](Caching/Caching.md)                                       |         |            |      |
| grpcWeb           | Enable gRPC Web                                                                    |  false  |            |      |
| proxyProtocol     | Enable proxy protocol                                                              |  true   |            |      |