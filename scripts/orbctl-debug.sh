#!/bin/bash

set -e

go run -race ./cmd/gen-executables/*.go \
  --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") \
  --commit $(git rev-parse HEAD) \
  --githubclientid "${ORBOS_GITHUBOAUTHCLIENTID}" \
  --githubclientsecret "${ORBOS_GITHUBOAUTHCLIENTSECRET}" \
  --orbctl ./artifacts \
  --debug \
  --dev
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
dlv exec --api-version 2 --headless --listen 127.0.0.1:2345 ./artifacts/orbctl-Linux-x86_64 -- "$@"
