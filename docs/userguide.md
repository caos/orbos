# Userguides

## Reset Cluster

If you would like to reset a cluster, in this example `kubernetes`, just follow this steps.

> !!! Danger Zone !!! This will reset all nodes including the `master` and `etcd` ones.

Delete on-cluster `Orbiter`:

```bash
kubectl -n caos-system delete deployment orbiter
```

Delete the secret `kubeconfig` from you git project

Start Orbiter localy:

```bash
orbctl -f [path to orb file] takeoff
```
