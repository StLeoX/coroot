FROM golang:1.23-bullseye AS backend-builder
RUN apt update && apt install -y liblz4-dev
WORKDIR /tmp/src
COPY go.mod .
COPY go.sum .
RUN export GOPROXY='https://goproxy.cn' && go mod download
COPY . .
ARG VERSION=latest
RUN export COROOT_VERSION=$VERSION && make go-build


FROM debian:bullseye
RUN apt update && apt install -y ca-certificates

COPY --from=backend-builder /tmp/src/coroot /usr/bin/coroot

VOLUME /data
EXPOSE 8888
EXPOSE 8889

ENTRYPOINT ["coroot"]
