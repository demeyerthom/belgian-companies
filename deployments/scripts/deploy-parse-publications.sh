#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd "${dir}"/../../ || exit

## Build Go binary
make build-parse-publications

## Build docker image
docker build --no-cache -t demeyerthom/parse-publications:latest -f ./deployments/docker/Dockerfile.parse-publications ./bin

# Stop previous container
docker stop parse-publications || true && docker rm parse-publications || true

# Run new one
docker run --name parse-publications -d \
    --network=production \
    --restart always \
    -e BROKERS=broker:9092 \
    demeyerthom/parse-publications:latest

## Remove binaries
make clean