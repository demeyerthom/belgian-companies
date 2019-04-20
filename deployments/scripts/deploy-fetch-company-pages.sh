#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd ${dir}/../../

## Build Go binary
//make build-fetch-company-pages

## Build docker image
docker build -t demeyerthom/fetch-company-pages:latest \
    -f ./deployments/docker/Dockerfile.fetch-company-pages \
    ./bin

# Stop previous container
docker stop fetch-company-pages || true && docker rm fetch-company-pages || true

# Run new one
docker run --name fetch-company-pages -d \
    --network=production \
    -e BROKERS=broker:9092 \
    -e PROXY_URL=socks5://proxy:9150 \
    -e PATH=/ \
    demeyerthom/fetch-company-pages:latest