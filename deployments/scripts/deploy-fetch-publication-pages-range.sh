#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd ${dir}/../../

make vendor

## Build Go binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

## Build docker image
docker build -t demeyerthom/fetch-publication-pages:latest \
    -f ./deployments/docker/Dockerfile.fetch-publication-pages \
    ./bin

docker save demeyerthom/fetch-publication-pages:latest | ssh -C root@192.168.178.37 docker load

# Stop previous container
ssh -C root@192.168.178.37 docker stop fetch-publication-pages-range || true
ssh -C root@192.168.178.37 docker rm fetch-publication-pages-range || true

# Run new one
ssh -C root@192.168.178.37 docker create --name fetch-publication-pages-range \
    --restart always \
    -e BROKERS=broker:9092 \
    -e PROXY_URL=socks5://proxy:9150 \
    -e START_DATE=${START_DATE:-2019-01-01} \
    -e END_DATE=${END_DATE} \
    -e PATH=/ \
    demeyerthom/fetch-publication-pages:latest \
    fetch-publication-pages range

ssh -C root@192.168.178.37 docker network connect kafka fetch-publication-pages-range
ssh -C root@192.168.178.37 docker network connect proxy fetch-publication-pages-range

ssh -C root@192.168.178.37 docker start fetch-publication-pages-range

## Remove binaries
make clean