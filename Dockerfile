# syntax=docker/dockerfile:1
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0
# Copyright (c) The OpenSoba Authors

# OpenSoba container image.
#
# Build for the local platform:
#   docker build -t opensoba .
#
# Build and push a multi-arch image to GHCR:
#   docker buildx build \
#     --platform linux/amd64,linux/arm64 \
#     --build-arg VERSION=1.16.4 \
#     -t ghcr.io/opensoba/opensoba:1.16.4 \
#     --push .

# ---------------------------------------------------------------------------
# Build stage
# ---------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS build

# TARGETOS/TARGETARCH are provided automatically by BuildKit when --platform
# is used, and used here to cross-compile a static binary.
ARG TARGETOS TARGETARCH
ARG VERSION=dev
ARG GIT_COMMIT=unknown

WORKDIR /src

# The repository is split into several replace-based modules (api, sdk,
# proto-public, envoyextensions, troubleshoot), so the whole tree is needed.
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath \
      -ldflags "-X github.com/opensoba/opensoba/version.GitCommit=${GIT_COMMIT} -X github.com/opensoba/opensoba/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      -o /out/soba .

# ---------------------------------------------------------------------------
# Runtime stage
# ---------------------------------------------------------------------------
FROM alpine:3.19 AS release

ARG VERSION=dev

LABEL org.opencontainers.image.authors="OpenSoba Team <opensoba@users.noreply.github.com>" \
      org.opencontainers.image.url="https://github.com/opensoba/opensoba" \
      org.opencontainers.image.documentation="https://github.com/opensoba/opensoba#readme" \
      org.opencontainers.image.source="https://github.com/opensoba/opensoba" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.vendor="OpenSoba" \
      org.opencontainers.image.title="soba" \
      org.opencontainers.image.description="OpenSoba is a datacenter runtime that provides service discovery, configuration, and orchestration." \
      org.opencontainers.image.licenses="MPL-2.0"

# libc6-compat is needed to symlink the shared libraries for ARM builds.
RUN apk add --no-cache \
        ca-certificates \
        dumb-init \
        libc6-compat \
        libcap \
        su-exec \
        tzdata \
 && addgroup -S soba \
 && adduser -S -G soba -h /home/soba -s /sbin/nologin soba

COPY --from=build /out/soba /bin/soba

# /soba/data holds agent state (persist it via a volume); /soba/config is the
# directory the agent is started with so extra config files can be dropped in.
RUN mkdir -p /soba/data /soba/config \
 && chown -R soba:soba /soba

# Set up nsswitch.conf for Go's "netgo" resolver, otherwise DNS supersedes the
# container's hosts file, which we don't want.
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

VOLUME /soba/data

# Server RPC (clients <-> servers), Serf LAN/WAN gossip, HTTP and DNS.
EXPOSE 8300 8301 8301/udp 8302 8302/udp 8500 8600 8600/udp

# OpenSoba doesn't need root, so the entrypoint drops privileges to the "soba"
# user. It also uses dumb-init as PID 1 to reap zombie processes.
COPY .release/docker/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]

# By default you get an insecure single-node development server that stores
# everything in RAM, exposes a web UI and HTTP endpoints, and bootstraps
# itself. Don't use this configuration for production.
CMD ["agent", "-dev", "-client", "0.0.0.0"]
