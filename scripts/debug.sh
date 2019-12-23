#!/bin/bash

set -e

JUMPHOST=$1

mkdir -p /tmp/orbctldev
go run ./cmd/gen-executables/*.go --debug --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") --commit $(git rev-parse HEAD) --orbctl /tmp/orbctldev

ssh ubuntu@${JUMPHOST} 'kill $(pgrep dlv) 2&> /dev/null || echo "no delve PID to kill"'
scp /tmp/orbctldev/orbctl-Linux-x86_64 ubuntu@${JUMPHOST}:/usr/local/bin/orbctl
ssh ubuntu@${JUMPHOST} "dlv exec /usr/local/bin/orbctl --api-version 2 --headless --listen 0.0.0.0:5000 --continue --accept-multiclient -- -f /home/ubuntu/.orb/test takeoff"
