# LabelSelector 
 

## Structure 
 

| Attribute        | Description                                                                                                                                                                                                                                                                | Default | Collection | Map  |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| matchLabels      | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed. +optional  |         |            | X    |
| matchExpressions | matchExpressions is a list of label selector requirements. The requirements are ANDed. +optional , [here](LabelSelectorRequirement/LabelSelectorRequirement.md)                                                                                                            |         | X          |      |