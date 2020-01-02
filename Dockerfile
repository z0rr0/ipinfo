FROM alpine:latest
MAINTAINER Alexander Zaytsev "me@axv.email"
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates tzdata
ADD ipinfo /bin/ipinfo
RUN chmod 0755 /bin/ipinfo
EXPOSE 8082
VOLUME ["/data/conf/", "/data/db/"]
ENTRYPOINT ["ipinfo"]
CMD ["-config", "/data/conf/ipinfo.json"]