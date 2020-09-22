# Using BOOM

## Setup the Orb

This step is only necessary if no already done for Orbiter or Zitadel-Operator.
To setup the configuration for an Orb, please follow [this guide](../orb/orb.md).

## Configure BOOM

To configure `BOOM`, a file with the name `boom.yml` has to be existent in the root directory of the Git repository.
The [example](../../examples/boom/boom.yml) can be used as basis, and so has to be copied to the root of the Git repository. 

## Structure of the used boom.yml 

The structure is documented in v1beta1 [here](yml/v1beta1/ToolsetSpec.md) and v1beta2 [here](yml/v1beta2/ToolsetSpec.md), from there you can follow the file-tree to what configurations you want to make.

## Deploy BOOM on Orb

If the Kubernetes is bootstraped with ORBOS, then a `BOOM` instance can be started with:
```bash
orbctl -f $HOME/.orb/config takeoff
```

Otherwise if Kubernetes is boostrapped any other way, then a `BOOM` instance can be started with:
```bash
orbctl -f $HOME/.orb/config takeoff --kubeconfig ${KUBECONFIG}
```