#!/usr/bin/env bash

#!/bin/bash
source ~/.bashrc

echo "Build Go binary in Docker image"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../bin/fetch-publication-pages ../cmd/fetch-publication-pages/

echo "Build docker image"
docker build -t belgian-companies/fetch-publication-pages:latest ../build/docker/Dockerfile.fetch-publication-pages

echo "Stop previous container"
docker stop belgian-companies-fetch-publication-pages || true && docker rm belgian-companies-fetch-publication-pages || true

echo "Run new one"
docker run --name belgian-companies-fetch-publication-pages -d \
#		-e ENV=production \
#		-e DB_HOST=mongodb://admin:pass@datastore:27017 \
#		-e DB_NAME=irs \
		belgian-companies/fetch-publication-pages:latest