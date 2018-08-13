FROM alpine:latest
MAINTAINER Alexander Zaytsev "thebestzorro@yandex.ru"
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates tzdata
ADD ipinfo /bin/ipinfo
RUN chmod 0755 /bin/ipinfo
EXPOSE 8082
VOLUME ["/data/conf/"]
ENTRYPOINT ["ipinfo"]
CMD ["-config", "/data/conf/ipinfo.json"]