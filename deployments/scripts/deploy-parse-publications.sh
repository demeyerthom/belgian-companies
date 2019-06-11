#!/bin/bash

#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd ${dir}/../../

make vendor

## Build Go binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publications ./cmd/parse-publications/

## Build docker image
docker build -t demeyerthom/parse-publications:latest \
    -f ./deployments/docker/Dockerfile.parse-publications \
    ./bin

docker save demeyerthom/parse-publications:latest | ssh -C root@192.168.178.37 docker load

# Stop previous container
ssh -C root@192.168.178.37 docker stop parse-publications || true
ssh -C root@192.168.178.37 docker rm parse-publications || true

# Run new one
ssh -C root@192.168.178.37 docker create --name parse-publications \
    --restart always \
    -e BROKERS=broker:9092 \
    demeyerthom/parse-publications:latest

ssh -C root@192.168.178.37 docker network connect kafka parse-publications

ssh -C root@192.168.178.37 docker start parse-publications

## Remove binaries
make clean