#!/bin/bash

set -ex

KEY=$1
DEST="orbiter@$2"
shift
shift

scp -i $KEY ./scripts/stop-na.sh $DEST:/usr/local/bin/stop-na.sh
scp -i $KEY ./scripts/debug-na.sh $DEST:/usr/local/bin/debug-na.sh
ssh -i $KEY $DEST debug-na.sh $@

