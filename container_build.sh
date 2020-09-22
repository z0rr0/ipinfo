#!/usr/bin/env bash

CONTAINER="golang:1.15-alpine"
SOURCES="${PWD}"
ATTRS="$(/bin/bash version.sh)"
IDCMD=$(command -v id)
DCMD=$(command -v docker)
PERM="$(${IDCMD} -u ${USER}):$(${IDCMD} -g ${USER})"

if [ -z "$DCMD" ]; then
  echo "docker not found"
  exit 1
fi

$DCMD run --rm --user "${PERM}" \
  --volume "${SOURCES}":/usr/app \
  --workdir /usr/app \
  --env GOCACHE=/tmp/.cache \
  ${CONTAINER} go build -v -ldflags "${ATTRS}"

if [[ $? -gt 0 ]]; then
  echo "ERROR: build container"
  exit 2
fi
