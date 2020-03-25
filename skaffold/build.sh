#!/bin/bash

set -e

export CGO_ENABLED=0

go run ./cmd/gen-executables/*.go --version skaffold --commit skaffold --debug $DEBUG --orbctl ./artifacts

TARGET=prod
if [[ "$DEBUG" == "true" ]]; then
  TARGET=build
fi

docker build --tag $IMAGE --build-arg DEBUG --target $TARGET .
[[ $PUSH_IMAGE=true ]] && docker push $IMAGE


