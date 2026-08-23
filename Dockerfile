ARG GOLANG_VERSION=1.27.0
ARG ALPINE_VERSION=3.24

# The builder runs natively on the host arch ($BUILDPLATFORM) and cross-compiles:
# with CGO_ENABLED=0 that is free, while an emulated builder would not be.
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /build

# download dependencies to cache them in a layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG GO_LDFLAGS=""
# TARGETOS/TARGETARCH come from BuildKit's --platform: `make docker` builds for
# the host arch, `make docker-push` for linux/amd64 and linux/arm64. Not named
# LDFLAGS: that one conventionally holds C linker flags and is often already
# exported by the host shell (Homebrew sets it).
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "$GO_LDFLAGS" -o /app/ipinfo .

# No --platform here: BuildKit resolves this to the target platform, so the
# binary above and this base image always match.
FROM alpine:${ALPINE_VERSION}
# No ca-certificates: the service only reads a local mmdb file, and time zones
# are embedded into the binary through the time/tzdata import.
RUN adduser -D -u 1000 appuser
LABEL org.opencontainers.image.authors="me@axv.email" \
    org.opencontainers.image.url="https://hub.docker.com/r/z0rr0/ipinfo" \
    org.opencontainers.image.documentation="https://github.com/z0rr0/ipinfo" \
    org.opencontainers.image.source="https://github.com/z0rr0/ipinfo" \
    org.opencontainers.image.licenses="BSD-3-Clause" \
    org.opencontainers.image.title="IPInfo" \
    org.opencontainers.image.description="IP info web service"

COPY --from=builder /app/ipinfo /bin/ipinfo
USER appuser
EXPOSE 8082
VOLUME ["/data/conf/"]
ENTRYPOINT ["/bin/ipinfo"]
CMD ["-config", "/data/conf/ipinfo.json"]
