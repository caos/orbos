#!/bin/bash

set -e

go run -race ./cmd/gen-executables/*.go --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") --commit $(git rev-parse HEAD) --orbctl /tmp/orbctldev
time /tmp/orbctldev/orbctl-Linux-x86_64 $@
