# ORBITER

## What Is It

`Orbiter` boostraps, lifecycles and destroys clustered software and other cluster managers whereas each can be configured to span over a wide range of infrastructure providers. Its focus is laid on automating away all `day two` operations, as we consider them to have much bigger impacts than `day one` operations from a business perspective.

## How Does It Work

An Orbiter instance runs as a Kubernetes Pod managing the configured clusters (i.e. an Orb), typically including the one it is running on. It scales the clusters nodes and has `Node Agents` install software packages on their operating systems. `Node Agents` run as native system processes managed by `systemd`. An Orbs Git repository is the only source of truth for desired state. Also, the current Orbs state is continously pushed to its Git repository, so not only changes to the desired state is always tracked but also the most important changes to the actual systems state.

For more details, take a look at the [design docs](orbiter/terminology.md).

## Why Another Cluster Manager

We observe a universal trend of increasing system distribution. Key drivers are cloud native engineering, microservices architectures, global competition among hyperscalers and so on.

We embrace this trend but counteract its biggest downside, the associated increase of complexity in managing all these distributed systems. Our goal is to enable players of any size to run clusters of any type using infrastructure from any provider. Orbiter is a tool to do this in a reliable, secure, auditable, cost efficient way, preventing vendor lock-in, monoliths consisting of microservices and human failure doing repetitive tasks.

What makes Orbiter special is that it ships with a nice **Mission Control UI** (currently in closed alpha) providing useful tools to interact intuitively with the operator. Also, the operational design follows the **GitOps pattern**, highlighting `day two operations`, sticking to a distinct source of truth for declarative system configuration and maintaining a consistent audit log, everything out-of-the-box. Then, the Orbiter code base is designed to be **highly extendable**, which ensures that any given cluster type can eventually run on any desired provider.

## How To Use It

In order to manage a Kubernetes cluster on already provisioned VMs, follow this guide.  

If you'd rather run a Kubernetes cluster on Google Compute Engine with ORBITER managed VMs, click here.

## Operating System Requirements

See [OS Requirements](orbiter/os-requirements.md) for details.

## Supported Clusters

See [Clusters](orbiter/clusters.md) for details.

## Supported Providers

See [Providers](orbiter/providers.md) for details.

## How To Contribute

See [contribute](orbiter/contribute.md) for details
