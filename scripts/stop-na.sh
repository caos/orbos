#!/bin/bash

set -x

sudo systemctl stop node-agentd
CODE="$?"
if [[ "$CODE" == "5" ]]; then
    exit 0
elif [[ "$CODE" != "0" ]]; then
    exit 1
fi
sudo kill $(pgrep dlv) 2> /dev/null || true
