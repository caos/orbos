#!/bin/bash

set -e

rm -rf ./artifacts/*

export CGO_ENABLED=0
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
go run ./cmd/gen-executables/*.go --version "${IMAGE}" --commit "$(git log --pretty=format:'%h' -n 1)" "--debug=$DEBUG" --orbctl ./artifacts --containeronly

TARGET=prod
if [[ "$DEBUG" == "true" ]]; then
  TARGET=build
fi

docker build --tag $IMAGE --target $TARGET --file ./build/orbos/Dockerfile .
[[ "$PUSH_IMAGE" == "true" ]] && docker push $IMAGE
rm -rf ./artifacts/*


