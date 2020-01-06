#!/bin/bash

set -e

JUMPHOST=$1

# creds
ssh ubuntu@${JUMPHOST} "mkdir -p /home/ubuntu/.orb && sudo chown -R 1000:1000 /usr/local/bin"
scp /home/elio/.orb/test ubuntu@${JUMPHOST}:/home/ubuntu/.orb/test

# delve
ssh ubuntu@${JUMPHOST} "sudo apt-get update && sudo apt-get install -y git && wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && sudo tar -zxvf go1.13.3.linux-amd64.tar.gz -C / && sudo chown -R $(id -u):$(id -g) /go && /go/bin/go get -u github.com/go-delve/delve/cmd/dlv && /go/bin/go install github.com/go-delve/delve/cmd/dlv && mv /home/ubuntu/go/bin/dlv /usr/local/bin/"
