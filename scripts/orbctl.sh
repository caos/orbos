#!/bin/bash

set -e

go run -race ./cmd/gen-executables/*.go \
  --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///")-dev \
  --commit $(git rev-parse HEAD) \
  --githubclientid "${ORBOS_GITHUBOAUTHCLIENTID}" \
  --githubclientsecret "${ORBOS_GITHUBOAUTHCLIENTSECRET}" \
  --orbctl ./artifacts \
  --dev 1>&2
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
time ./artifacts/orbctl-Linux-x86_64 "$@"
