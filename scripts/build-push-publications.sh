#!/usr/bin/env bash

echo "Build Go binary"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/push-publications ./cmd/push-publications/

echo "Build docker image"
docker build -t demeyerthom/push-publications:latest -f ./build/docker/Dockerfile.push-publications .

echo "Stop previous container"
docker stop belgian-companies-push-publications || true && docker rm belgian-companies-push-publications || true

echo "Run new one"
docker run --name belgian-companies-push-publications -d demeyerthom/push-publications:latest