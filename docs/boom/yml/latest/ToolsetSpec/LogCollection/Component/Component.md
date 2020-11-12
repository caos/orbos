# Component 
 

## Structure 
 

| Attribute    | Description                                                                    | Default | Collection | Map  |
| ------------ | ------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| nodeSelector | NodeSelector for deployment                                                    |         |            | X    |
| tolerations  | Tolerations to run fluentbit on nodes , [here](k8s/Tolerations/Tolerations.md) |         | X          |      |
| resources    | Resource requirements , [here](k8s/Resources/Resources.md)                     |         |            |      |