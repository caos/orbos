#!/usr/bin/env bash
publicKeyFile=$1
tmpFolder=$2

for MACHINE in  master1 worker1
do
    ./machine.sh $MACHINE ks.cfg ${publicKeyFile} ${tmpFolder} centos.iso
done
