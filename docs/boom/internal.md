# Internal logic

## Folder structure

The boom will extend the existing tools folder with different subfolders for the necessary applications.
For each crd there will be an subfolder for each application where the generated values.yaml are stored.
As a result of this files there will be an results.yaml under the subfolder *crd-name*/results/.

Like this:

* tools
  * logging-operator
    * *crd-name*
      * values.yaml
      * results
        * results.yaml
  * *helm*
  * *charts*

also are there the differnt tools for templating, the charts folder consists of all fetchet charts localy, and the helm folder is the helm-home folder. The charts will get fetched during the docker build phase with a seperate binary.

## toolsets

To add any new toolset or change existing ones look into the bundles folder under internal/bundle/bundles.
Every defined bundle gets a function in there and can be referenced with a name of type name.Bundle.

## used tools

The following cli-tools are used from the boom:

* helm
* kubectl
* kustomize(just for applying the toolset-crd)

As they are used, they also have to be installed into the image during the docker build.
