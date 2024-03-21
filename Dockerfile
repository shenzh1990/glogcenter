FROM alpine:latest
LABEL MAINTAINER="mllog"

WORKDIR /glc
COPY ./glc/bin/linux64 ./
COPY ./glc/www/web/dist ./web/dist

RUN chmod 755 ./glocenter

EXPOSE 8080

ENTRYPOINT ./glocenter