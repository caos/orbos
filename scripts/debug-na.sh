#!/bin/bash

sudo systemctl stop node-agentd

if go/bin/dlv version; then
    echo "Skipping Delve installation"
else
    curl -O  https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz
    tar -xf go1.13.8.linux-amd64.tar.gz
    ./go/bin/go get -u github.com/go-delve/delve/cmd/dlv
    go/bin/dlv version
fi

go/bin/dlv exec /usr/local/bin/node-agent --api-version 2 --headless --listen 0.0.0.0:5001 -- $@