#!/usr/bin/env bash
# getter for gopass and secret yaml creation
set -e
[[ `uname` = "Linux" ]] && ENCODE="base64 --wrap=0" || ENCODE="base64"

# apply via: secrets.yaml.sh | kubectl apply -f -

#update remote passwords
#gopass sync &> /dev/null


# get gopass secrets and output secret yaml
GITHUB_IMAGE_PULL_SECRET=$(cat ~/.docker/config.json | $ENCODE)
ORBCONFIG=$(cat ~/.orb/test | $ENCODE)

NAMESPACE=caos-system

cat <<EOL
---
apiVersion: v1
data:
  .dockerconfigjson: $GITHUB_IMAGE_PULL_SECRET
kind: Secret
metadata:
  name: gcr
  namespace: $NAMESPACE
type: kubernetes.io/dockerconfigjson
---
apiVersion: v1
data:
  orbconfig: $ORBCONFIG
kind: Secret
metadata:
  name: caos
  namespace: $NAMESPACE
type: Opaque
---
EOL
