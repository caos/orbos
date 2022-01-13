#!/bin/bash

ORBITER_BOOTSTRAP_PUBLICKEY="$(cat $2)" envsubst < $1 > /tmp/ks.cfg

virt-install \
    --virt-type kvm \
    --os-type linux \
    --os-variant rhel7 \
    --disk size=10 \
    --location 'http://mirror.init7.net/centos/7.9.2009/os/x86_64/' \
    --initrd-inject=/tmp/ks.cfg \
    --memory 4096 \
    --vcpus 2 \
    --nographics \
    --extra-args "console=ttyS0 ks=file:/ks.cfg" \
    --name $3
