# Spec 
 

## Structure 
 

| Attribute    | Description                                                                       | Default | Collection | Map  |
| ------------ | --------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| controlPlane | Configuration for the control plane for Kubernetes , [here](Pool/Pool.md)         |         |            |      |
| networking   | Configuration for the networking in kubernetes , [here](Networking/Networking.md) |         |            |      |
| verbose      | Flag to set log-level to debug                                                    |         |            |      |
| versions     | Versions to ensure for the components , [here](CompVersions/CompVersions.md)      |         |            |      |
| workers      | List of configurations for the worker pools , [here](Pool/Pool.md)                |         | X          |      |