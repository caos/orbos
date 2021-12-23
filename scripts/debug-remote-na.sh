#!/bin/bash

set -ex

KEY=$1
DEST="orbiter@$2"
shift
shift

scp -i $KEY ./scripts/stop-na.sh $DEST:/home/orbiter/stop-na-tmp.sh
ssh -i $KEY $DEST -- sudo mv /home/orbiter/stop-na-tmp.sh /usr/local/bin/stop-na.sh
ssh -i $KEY $DEST -- stop-na.sh
go run ./cmd/chore/dev/executables/*.go
scp -i $KEY ./artifacts/nodeagent $DEST:/home/orbiter/node-agent-tmp
ssh -i $KEY $DEST -- sudo mv /home/orbiter/node-agent-tmp /usr/local/bin/node-agent
scp -i $KEY ./scripts/debug-na.sh $DEST:/home/orbiter/debug-na-tmp.sh
scp -i $KEY $DEST -- sudo mv /home/orbiter/debug-na-tmp.sh /usr/local/bin/debug-na.sh
ssh -i $KEY $DEST debug-na.sh $(curl ifconfig.co/) "$@"

