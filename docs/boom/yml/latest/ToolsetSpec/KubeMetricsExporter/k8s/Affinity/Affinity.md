# Affinity 
 

## Structure 
 

| Attribute       | Description                                                                                                                                                                          | Default | Collection | Map  |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| nodeAffinity    | Describes node affinity scheduling rules for the pod. +optional , [here](NodeAffinity/NodeAffinity.md)                                                                               |         |            |      |
| podAffinity     | Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)). +optional , [here](PodAffinity/PodAffinity.md)                  |         |            |      |
| podAntiAffinity | Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)). +optional , [here](PodAntiAffinity/PodAntiAffinity.md) |         |            |      |