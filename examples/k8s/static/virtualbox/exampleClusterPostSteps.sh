#!/usr/bin/env bash

tmpFolder=$1

for MACHINE in master1 worker1
do
    ./poststepps.sh $MACHINE ${tmpFolder}
done