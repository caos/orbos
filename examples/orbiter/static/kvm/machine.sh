#!/bin/bash

ORBITER_BOOTSTRAP_PUBLICKEY="$(cat $2)" envsubst < $1 > /tmp/ks.cfg

virt-install \
    --virt-type kvm \
    --os-type linux \
    --os-variant ubuntu20.04 \
    --disk size=10 \
    --location 'http://mirror.init7.net/ubuntu-releases/20.04/ubuntu-20.04.4-live-server-amd64.iso' \
    --initrd-inject=/tmp/ks.cfg \
    --memory 4096 \
    --vcpus 2 \
    --nographics \
    --extra-args "console=ttyS0 ks=file:/ks.cfg" \
    --name $3
