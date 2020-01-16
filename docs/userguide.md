# Userguides

## Reset Cluster

If you would like to reset a cluster, in this example `kubernetes`, follow these steps.

> !!! Danger Zone !!! This will reset all nodes including the `master` and `etcd` nodes.

Delete on-cluster `Orbiter`:

```bash
kubectl -n caos-system delete deployment orbiter
```

Delete the secret `kubeconfig` from your git project

Start Orbiter locally:

```bash
orbctl -f [path to orb file] takeoff
```
