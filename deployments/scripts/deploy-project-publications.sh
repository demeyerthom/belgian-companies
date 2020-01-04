#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd "${dir}"/../../ || exit

## Build Go binary
make build-project-publications

## Build docker image
docker build --no-cache -t demeyerthom/project-publications:latest -f ./deployments/docker/Dockerfile.project-publications ./bin

# Stop previous container
docker stop project-publications || true && docker rm project-publications || true

# Run new one
docker run --name project-publications -d \
    --network=production \
    --restart always \
    -e BROKERS=broker:9092 \
    -e ELASTIC_URL=http://elasticsearch:9200 \
    demeyerthom/project-publications:latest