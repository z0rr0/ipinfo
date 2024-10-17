#!/usr/bin/env bash

TAG=$(git tag | sort -V | tail -1)
VERSION="${TAG:1}"

echo "version: ${VERSION}"

# add tag as latest version to new image
docker tag z0rr0/ipinfo:latest z0rr0/ipinfo:"${VERSION}"

# send images to docker hub
docker push z0rr0/ipinfo:"${VERSION}"
docker push z0rr0/ipinfo:latest
