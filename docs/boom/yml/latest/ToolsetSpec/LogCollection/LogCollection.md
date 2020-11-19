# LogCollection 
 

## Structure 
 

| Attribute       | Description                                                                                | Default | Collection | Map  |
| --------------- | ------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| deploy          | Flag if tool should be deployed                                                            |  false  |            |      |
| fluentd         | Fluentd Specs , [here](Fluentd/Fluentd.md)                                                 |         |            |      |
| fluentbit       | Fluentbit Specs , [here](Component/Component.md)                                           |         |            |      |
| operator        | Logging operator Specs , [here](Component/Component.md)                                    |         |            |      |
| clusterOutputs  | ClusterOutputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified  |         | X          |      |
| outputs         | Outputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified         |         | X          |      |
| watchNamespaces | Watch these namespaces                                                                     |         | X          |      |