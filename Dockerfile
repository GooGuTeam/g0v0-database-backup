FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o backup-watcher .

FROM ubuntu:24.04

WORKDIR /app

RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    lsb-release \
    percona-xtrabackup \
    zstd \
    mysql-client \
    ca-certificates \
    unzip \
    rclone \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/backup-watcher /app/backup-watcher
RUN chmod +x /app/backup-watcher

COPY ./config.json /app/config.json
COPY ./rclone.conf /app/rclone.conf

EXPOSE 32400

HEALTHCHECK --interval=5m --timeout=10s --start-period=10s \
    CMD curl -f http://localhost:32400/health || exit 1

VOLUME ["/backup", "/data", "/downloaded_backup"]

CMD ["/app/backup-watcher"]
