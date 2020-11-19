#!/bin/bash

set -e

go run -race ./cmd/gen-executables/*.go \
  --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") \
  --commit $(git rev-parse HEAD) \
  --githubclientid "${ORBOS_GITHUBOAUTHCLIENTID}" \
  --githubclientsecret "${ORBOS_GITHUBOAUTHCLIENTSECRET}" \
  --orbctl ./artifacts \
  --debug \
  --dev \
  "$@"
