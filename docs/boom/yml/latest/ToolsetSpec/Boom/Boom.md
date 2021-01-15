# Boom 
 

## Structure 
 

| Attribute           | Description                                                               | Default  | Collection | Map  |
| ------------------- | ------------------------------------------------------------------------- | -------- | ---------- | ---  |
| version             | Version of BOOM which should be reconciled                                |          |            |      |
| nodeSelector        | NodeSelector for boom deployment                                          |          |            | X    |
| tolerations         | Tolerations to run boom on nodes , [here](k8s/Tolerations/Tolerations.md) |          | X          |      |
| resources           | Resource requirements , [here](k8s/Resources/Resources.md)                |          |            |      |
| customImageRegistry | Use this registry to pull the BOOM image from                             |  ghcr.io |            |      |