#!/usr/bin/env bash

GITCMD=$(command -v git)
if [ -z "$GITCMD" ]; then
  echo "git command was not found"
  exit 2
fi
TS="$(/bin/date -u +\"%F_%T\")UTC"
TAG=$(${GITCMD} tag | sort --version-sort | tail -1)
VER=$(${GITCMD} log --oneline | head -1)

if [[ -z "$TAG" ]]; then
  TAG="N/A"
fi
echo "-X main.Version=${TAG} -X main.Revision=git:${VER:0:7} -X main.BuildDate=${TS}"
