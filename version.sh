#!/usr/bin/env bash

cd ${GOPATH}/src/github.com/z0rr0/ipinfo

FILENAME="version"
TS="`TZ=UTC date +\"%F_%T\"`UTC"
TAG="`git tag | sort --version-sort | tail -1`"
VER="`git log --oneline | head -1 `"

if [[ -z "$TAG" ]]; then
    TAG="N/A"
fi

echo "-X main.Version=${TAG} -X main.Revision=git:${VER:0:7} -X main.BuildDate=${TS}"