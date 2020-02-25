#!/bin/bash

sudo systemctl stop node-agentd || exit 1

if ./go/bin/dlv version; then
    echo "Skipping Delve installation"
else
    curl -O  https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz || exit 1
    tar -xf go1.13.8.linux-amd64.tar.gz || exit 1
    sudo yum install -y git
    ./go/bin/go get -u github.com/go-delve/delve/cmd/dlv || exit 1
    ./go/bin/dlv version || exit 1
fi

sudo firewall-cmd --permanent --add-port 5001/tcp || exit 1
sudo firewall-cmd --reload || exit 1

sudo kill $(pgrep dlv) 2> /dev/null

exec sudo ./go/bin/dlv exec /usr/local/bin/node-agent --api-version 2 --headless --listen 0.0.0.0:5001 -- --ignore-ports 5001 $@