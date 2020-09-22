#!/usr/bin/env bash

CONTAINER="golang:1.15-alpine"
SOURCES="${GOPATH}/src"
TARGET="${GOPATH}/bin/alpine"
ATTRS="$(/bin/bash version.sh)"
IDCMD=$(command -v id)
DCMD=$(command -v docker)
PERM="$(${IDCMD} -u ${USER}):$(${IDCMD} -g ${USER})"

if [ -z "$DCMD" ]; then
  echo "docker not found"
  exit 1
fi

rm -rf "${TARGET}"
mkdir -p "${TARGET}"/bin "${TARGET}"/pkg

$DCMD run --rm --user "${PERM}" \
  --volume "${SOURCES}":/usr/p/src:ro \
  --volume "${TARGET}"/pkg:/usr/p/pkg \
  --volume "${TARGET}"/bin:/usr/p/bin \
  --workdir /usr/p/src/github.com/z0rr0/ipinfo \
  --env GOPATH=/usr/p \
  --env GOCACHE=/tmp/.cache \
  ${CONTAINER} go install -v -ldflags "${ATTRS}" github.com/z0rr0/ipinfo

if [[ $? -gt 0 ]]; then
  echo "ERROR: build container"
  exit 2
fi

cp -v "${TARGET}"/bin/ipinfo ./
