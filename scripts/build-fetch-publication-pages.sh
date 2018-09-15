#!/usr/bin/env bash

echo "Build Go binary"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

echo "Build docker image"
docker build -t demeyerthom/fetch-publication-pages:latest -f ./build/docker/Dockerfile.fetch-publication-pages .

echo "Stop previous container"
docker stop belgian-companies-fetch-publication-pages || true && docker rm belgian-companies-fetch-publication-pages || true

echo "Run new one"
docker run --name belgian-companies-fetch-publication-pages -d demeyerthom/fetch-publication-pages:latest