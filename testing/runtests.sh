#!/usr/bin/env bash

npm install mochawesome  --save-dev

docker run \
-it -v $PWD:/e2e -w /e2e cypress/included:8.0.0 

