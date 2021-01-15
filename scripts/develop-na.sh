#!/bin/bash

set -ex

KEY=$1
DEST="orbiter@$2"

./scripts/build-debug-bins.sh --host-bins-only --commit debug 
scp -i $KEY ./scripts/stop-na.sh $DEST:/usr/local/bin/stop-na.sh
ssh -i $KEY $DEST -- stop-na.sh
scp -i $KEY ./artifacts/nodeagent $DEST:/usr/local/bin/node-agent
./scripts/debug-remote-na.sh "$@"
