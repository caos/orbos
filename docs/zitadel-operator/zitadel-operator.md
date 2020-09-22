# Zitadel-Operator

## What Is It

`Zitadel-Operator` bootstraps, lifecycles and destroy resources on Kubernetes so that our Cloud Native Identity and Access Management solution `Zitadel`,
more information [found here](https://github.com/caos/zitadel), runs on the Orb. 

## How Does It Work

The Operator ensures all resources needed to run `Zitadel` on the Orb and outside so that connections to `Zitadel` are possible. An Orb Git repository is used as only source of truth for the desired state. The desired state is being constantly compared to the current state on the orb to calculate what actions have to be taken.

## How To Use It

In order to see how to use the Zitadel-Operator in combination with Orbos or on a standalone Kubernetes-cluster, follow the instructions [here](./setup.md).