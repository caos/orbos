#!/bin/bash

set -ex

sudo systemctl stop node-agentd || exit 1
sudo kill $(pgrep dlv) 2> /dev/null || true
