#!/bin/bash

envsubst < $PWD/promtail/config.template.yaml > $PWD/promtail/config.yaml

docker run --rm --name promtail --volume "$PWD/promtail:/etc/promtail" --volume "$PWD/log:/var/log" grafana/promtail:master -config.file=/etc/promtail/config.yaml
