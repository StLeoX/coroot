FROM golang:1.23-bullseye AS backend-builder
RUN apt update && apt install -y liblz4-dev
WORKDIR /tmp/src
COPY go.mod .
COPY go.sum .
RUN export GOPROXY='https://goproxy.cn' && go mod download
COPY . .
ARG VERSION=v1.24.1-fix
RUN go build -mod=readonly -ldflags "-X main.version=$VERSION" -o coroot .


FROM debian:bullseye
RUN apt update && apt install -y ca-certificates

COPY --from=backend-builder /tmp/src/coroot /usr/bin/coroot

VOLUME /data
EXPOSE 8888

ENTRYPOINT ["coroot"]
