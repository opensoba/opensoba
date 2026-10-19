<h1>
  <img src="./.github/assets/logo-opensoba.svg" align="left" height="46px" alt="OpenSoba logo"/>
  <span>OpenSoba</span>
</h1>

[![License: MPL-2.0](https://img.shields.io/badge/License-MPL--2.0-brightgreen.svg)](LICENSE)

OpenSoba is a distributed, highly available, datacenter-aware solution to
connect and configure applications across dynamic, distributed infrastructure.

> ### About this project
>
> OpenSoba is an **independent community fork of HashiCorp Consul**, based on
> **Consul v1.16.4** — the last release that HashiCorp published under the
> Mozilla Public License 2.0 (MPL-2.0). The fork exists to keep that
> MPL-licensed codebase available and maintained as open source.
>
> OpenSoba is **not affiliated with, endorsed by, or sponsored by HashiCorp**.
> "HashiCorp" and "Consul" are trademarks of HashiCorp, Inc. and are used here
> only to identify and credit the upstream project.
>
> The command-line binary is named **`soba`**.

OpenSoba provides several key features:

* **Multi-Datacenter** — OpenSoba is built to be datacenter aware, and can
  support any number of regions without complex configuration.

* **Service Mesh** — OpenSoba Service Mesh enables secure service-to-service
  communication with automatic TLS encryption and identity-based authorization.
  Applications can use sidecar proxies in a service mesh configuration to
  establish TLS connections for inbound and outbound connections with
  Transparent Proxy.

* **Service Discovery** — OpenSoba makes it simple for services to register
  themselves and to discover other services via a DNS or HTTP interface.
  External services such as SaaS providers can be registered as well.

* **Health Checking** — Health Checking enables OpenSoba to quickly alert
  operators about any issues in a cluster. The integration with service
  discovery prevents routing traffic to unhealthy hosts and enables service
  level circuit breakers.

* **Dynamic App Configuration** — An HTTP API that allows users to store
  indexed objects, like configuration parameters and application metadata,
  within OpenSoba.

* **External Sources** — A service that registers itself with an
  `external-source` metadata value is shown in the web UI with the logo of
  that product. OpenBao, OpenTofu, OpenWonton and OpenDumpling are recognised
  out of the box, and [docs/dev](./docs/dev/README.md) registers a demo
  service for each of them.

* **SSO Login (OIDC)** — Users can log in to the web UI and the `soba` CLI
  through an OpenID Connect provider (for example Keycloak, Okta, Auth0, or
  Google) using the `oidc` auth method type. See
  [docs/oidc.md](./docs/oidc.md).

OpenSoba runs on Linux, macOS, FreeBSD, Solaris, and Windows and includes an
optional browser based UI.

## Quick Start

Start a single-node development agent (everything in memory, bootstrap a
cluster, expose the HTTP API and web UI):

```shell
soba agent -dev
```

Then explore the cluster:

```shell
soba members
soba catalog services
soba version
```

## Building from source

You need [Go](https://go.dev) 1.20+ to build the `soba` binary:

```shell
go build -o soba .
./soba version
```

## Releases and container image

Prebuilt `soba` archives for the same OS/arch set that upstream Consul v1.16.4
shipped (darwin, freebsd, linux, solaris, windows) are attached to every
[release](https://github.com/opensoba/opensoba/releases), together with a
`SHA256SUMS` file.

A multi-arch container image is published to the GitHub Container Registry:

```shell
docker run --rm -p 8500:8500 ghcr.io/opensoba/opensoba:latest
```

See [RELEASING.md](./RELEASING.md) for how releases are produced.

## Documentation

Upstream Consul documentation is a useful reference for the concepts and the
HTTP/DNS APIs, which OpenSoba keeps compatible. Where this fork has diverged in
naming, the `soba` binary and this repository are authoritative.

## Contributing

See [`.github/CONTRIBUTING.md`](./.github/CONTRIBUTING.md).

Because OpenSoba is built on the last MPL-2.0 release, **do not** contribute or
merge code copied from any HashiCorp release published under the Business
Source License (BUSL-1.1), i.e. any Consul release after v1.16.4. See
[`NOTICE.md`](./NOTICE.md) for details.

## License

OpenSoba is distributed under the [Mozilla Public License 2.0](./LICENSE)
(MPL-2.0), the same license as the upstream code it is forked from. See
[`NOTICE.md`](./NOTICE.md) for copyright and attribution details.

## Security

Please report security issues privately through the **Security** tab of this
repository (GitHub private vulnerability reporting). Do not report OpenSoba
issues to HashiCorp.
