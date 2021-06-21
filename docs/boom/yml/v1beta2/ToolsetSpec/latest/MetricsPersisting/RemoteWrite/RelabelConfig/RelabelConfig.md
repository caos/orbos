# RelabelConfig 
 

## Structure 
 

| Attribute    | Description                                                                                                                                                                                                        | Default | Collection | Map  |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| sourceLabels | The source labels select values from existing labels. Their content is concatenated using the configured separator and matched against the configured regular expression for the replace, keep, and drop actions.  |         | X          |      |
| separator    | Separator placed between concatenated source label values. default is ';'.                                                                                                                                         |         |            |      |
| targetLabel  | Label to which the resulting value is written in a replace action. It is mandatory for replace actions. Regex capture groups are available.                                                                        |         |            |      |
| regex        | Regular expression against which the extracted value is matched. Default is '(.*)'                                                                                                                                 |         |            |      |
| modulus      | Modulus to take of the hash of the source label values.                                                                                                                                                            |         |            |      |
| replacement  | Replacement value against which a regex replace is performed if the regular expression matches. Regex capture groups are available. Default is '$1'                                                                |         |            |      |
| action       | Action to perform based on regex matching. Default is 'replace'                                                                                                                                                    |         |            |      |