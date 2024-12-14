ARG GOLANG_VERSION="1.23.4"

FROM golang:${GOLANG_VERSION}-alpine AS builder
ARG LDFLAGS
WORKDIR /go/src/github.com/z0rr0/ipinfo
COPY . .
RUN echo "LDFLAGS = $LDFLAGS"
RUN GOOS=linux go build -ldflags "$LDFLAGS" -o ./ipinfo

FROM alpine:3.21
LABEL org.opencontainers.image.authors="me@axv.email" \
        org.opencontainers.image.url="https://hub.docker.com/r/z0rr0/ipinfo" \
        org.opencontainers.image.documentation="https://github.com/z0rr0/ipinfo" \
        org.opencontainers.image.source="https://github.com/z0rr0/ipinfo" \
        org.opencontainers.image.licenses="BSD-3-Clause" \
        org.opencontainers.image.title="IPInfo" \
        org.opencontainers.image.description="IP info web service"
COPY --from=builder /go/src/github.com/z0rr0/ipinfo/ipinfo /bin/
RUN chmod 0755 /bin/ipinfo

EXPOSE 8082
VOLUME ["/data/conf/"]
ENTRYPOINT ["ipinfo"]
CMD ["-config", "/data/conf/ipinfo.json"]
