#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd "${dir}"/../../ || exit

## Build Go binary
make build-fetch-publication-pages

## Build docker image
docker build -t demeyerthom/fetch-publication-pages:latest \
    -f ./deployments/docker/Dockerfile.fetch-publication-pages \
    ./bin

# Stop previous container
docker stop fetch-publication-pages-cron || true && docker rm fetch-publication-pages-cron || true

# Run new one
docker run --name fetch-publication-pages-cron -d \
    --restart always \
    --network=production \
    -e BROKERS=broker:9092 \
    -e PROXY_URL=socks5://proxy:9150 \
    -e PATH=/ \
    demeyerthom/fetch-publication-pages:latest \
    fetch-publication-pages cron

## Remove binaries
make clean