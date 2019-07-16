FROM golang:1.12 AS build
ARG APP_VER
WORKDIR /stem
ADD http://download.cdn.yandex.net/mystem/mystem-3.1-linux-64bit.tar.gz ./
RUN tar -xzf mystem-3.1-linux-64bit.tar.gz
COPY *.go ./
RUN CGO_ENABLED=0 go build -o app -ldflags "-X main.version=$APP_VER -s -w" ./

FROM ubuntu:18.04
RUN groupadd -r stem && useradd --no-log-init -r -g stem stem
WORKDIR /stem
RUN mkdir dict && chown stem:stem dict
COPY --chown=stem:stem --from=build /stem/mystem ./
COPY --chown=stem:stem --from=build /stem/app ./

USER stem
EXPOSE 8080
ENTRYPOINT ["/stem/app"]