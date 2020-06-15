# BOOM: the base tooling operator

## What is it

`BOOM` is designed to ensure that someone can create a reproducable "platform" with tools which are tested for their interoperability.

Currently we include the following tools:

- Ambassador Edge Stack
- Prometheus Operator
- Grafana
- logging-operator
- kube-state-metrics
- prometheus-node-exporter
- prometheus-systemd-exporter
- loki
- ArgoCD

Upcoming tools:

- Flux

## How does it work

The operator works by reading a configuration (crd) located in a GIT Repository.
In our default setup our "cluster lifecycle" tool `ORBITER`, shares the repository and secrets with `BOOM`.

```yaml
apiVersion: boom.caos.ch/v1beta1
kind: Toolset
metadata:
  name: caos
  namespace: caos-system
spec:
  preApply:
    deploy: true
    folder: preapply
  postApply:
    deploy: true
    folder: postapply
  prometheus-operator:
    deploy: true
  logging-operator:
    deploy: true
  prometheus-node-exporter:
    deploy: true
  grafana:
    deploy: true
  ambassador:
    deploy: true
    service:
      type: LoadBalancer
  kube-state-metrics:
    deploy: true
  prometheus:
    deploy: true
    storage:
      size: 5Gi
      storageClass: standard
  loki:
    deploy: true
    storage:
      size: 5Gi
      storageClass: standard
```

## How to use it

There has to be a git-repository with an boom.yml in the base. Then a `BOOM` instance can be started with 
```bash
orbctl -f $HOME/.orb/config takeoff
```

## Structure of the used boom.yml 

The structure is documented [here](yml/ToolsetSpec.md), from there you can follow the file-tree to what configurations you want to make.
