# LabelSelectorRequirement 
 

## Structure 
 

| Attribute | Description                                                                                                                                                                                                                                           | Default | Collection | Map  |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| key       | key is the label key that the selector applies to. +patchMergeKey=key +patchStrategy=merge                                                                                                                                                            |         |            |      |
| operator  | operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.                                                                                                                                  |         |            |      |
| values    | values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch. +optional  |         | X          |      |