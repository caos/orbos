#!/bin/bash

DEST=$1
shift

scp ./scripts/debug-na.sh $DEST:/usr/local/bin/debug-na.sh
ssh $DEST debug-na.sh $@
