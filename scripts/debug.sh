#!/bin/bash

set -e

go run ./cmd/gen-executables/*.go --debug --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") --commit $(git rev-parse HEAD) --orbctl /tmp/orbctldev 
dlv exec --api-version 2 --headless --listen 127.0.0.1:5000 /tmp/orbctldev/orbctl-Linux-x86_64 -- -f ${1} takeoff --recur --deploy=false
