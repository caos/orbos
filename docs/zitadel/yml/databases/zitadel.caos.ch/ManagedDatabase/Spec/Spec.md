# Spec 
 

## Structure 
 

| Attribute       | Description                                                                                    | Default | Collection | Map  |
| --------------- | ---------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| verbose         | Verbose flag to set debug-level to debug                                                       |         |            |      |
| replicaCount    | Number of replicas for the cockroachDB statefulset                                             |         |            |      |
| storageCapacity | Capacity for the PVC for cockroachDB                                                           |         |            |      |
| storageClass    | Storageclass for the PVC for cockroachDB                                                       |         |            |      |
| nodeSelector    | Nodeselector for statefulset and migration jobs                                                |         |            | X    |
| clusterDNS      | DNS entry used for cockroachDB certificates, should be the same use for the cluster-DNS-entry  |         |            |      |