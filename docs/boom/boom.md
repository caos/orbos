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

The operator works by reading a configuration (crd) located in a GIT Repository and 
calculates what actions have to be taken to ensure the desired state in the Git repository on the cluster.

In our default setup our "cluster lifecycle" tool `ORBITER`, shares the repository and secrets with `BOOM`.

## How to use it

In order to see how to use `BOOM` in combination with `Orbos` or on a standalone Kubernetes-cluster, follow the instructions [here](./setup.md).