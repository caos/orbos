# WeightedPodAffinityTerm 
 

## Structure 
 

| Attribute       | Description                                                                                                           | Default | Collection | Map  |
| --------------- | --------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| weight          | weight associated with matching the corresponding podAffinityTerm, in the range 1-100.                                |         |            |      |
| podAffinityTerm | Required. A pod affinity term, associated with the corresponding weight. , [here](PodAffinityTerm/PodAffinityTerm.md) |         |            |      |