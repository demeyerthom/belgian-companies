#!/usr/bin/env bash

echo "Build Go binary"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publication-pages ./cmd/parse-publication-pages/

echo "Build docker image"
docker build -t demeyerthom/parse-publication-pages:latest -f ./build/docker/Dockerfile.parse-publication-pages .

echo "Stop previous container"
docker stop belgian-companies-parse-publication-pages || true && docker rm belgian-companies-parse-publication-pages || true

echo "Run new one"
docker run --name belgian-companies-parse-publication-pages -d demeyerthom/parse-publication-pages:latest