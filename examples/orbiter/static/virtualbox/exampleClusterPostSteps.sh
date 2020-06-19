#!/usr/bin/env bash

tmpFolder=$1

for MACHINE in master1 worker1
do
    ./poststeps.sh $MACHINE ${tmpFolder}
done