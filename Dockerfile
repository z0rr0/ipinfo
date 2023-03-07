ARG GOLANG_VERSION="1.20.1"

FROM golang:${GOLANG_VERSION}-alpine as builder
ARG LDFLAGS
RUN apk --no-cache add ca-certificates tzdata git sqlite gcc libc-dev
WORKDIR /go/src/github.com/z0rr0/ipinfo
COPY . .
RUN echo "LDFLAGS = $LDFLAGS"
RUN GOOS=linux go build -ldflags "$LDFLAGS" -o ./ipinfo

FROM alpine:3.17
MAINTAINER Alexander Zaitsev "me@axv.email"
COPY --from=builder /go/src/github.com/z0rr0/ipinfo/ipinfo /bin/
RUN chmod 0755 /bin/ipinfo

EXPOSE 8082
VOLUME ["/data/conf/"]
ENTRYPOINT ["ipinfo"]
CMD ["-config", "/data/conf/ipinfo.json"]
