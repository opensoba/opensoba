# Releasing OpenSoba

Releases are fully automatic: **push a tag, get all the assets.**

## Steps

1. Bump the version (optional but recommended so `main` and the tag agree):

   ```sh
   echo "1.16.5" > version/VERSION
   git commit -am "Bump version to 1.16.5"
   git push
   ```

2. Tag and push the tag:

   ```sh
   git tag v1.16.5
   git push origin v1.16.5
   ```

That's it. The `Release` workflow then:

- cross-compiles `soba` for every OS/arch that upstream Consul v1.16.4 shipped,
  and publishes them as a GitHub Release:
  - `soba_<version>_darwin_amd64.zip`
  - `soba_<version>_darwin_arm64.zip`
  - `soba_<version>_freebsd_386.zip`
  - `soba_<version>_freebsd_amd64.zip`
  - `soba_<version>_linux_386.zip`
  - `soba_<version>_linux_amd64.zip`
  - `soba_<version>_linux_arm.zip`
  - `soba_<version>_linux_arm64.zip`
  - `soba_<version>_solaris_amd64.zip`
  - `soba_<version>_windows_386.zip`
  - `soba_<version>_windows_amd64.zip`
  - `soba_<version>_SHA256SUMS`
- builds a multi-arch (linux/amd64 + linux/arm64) container image and pushes it to
  `ghcr.io/opensoba/opensoba:<version>` and `:latest`.

The version embedded in the binaries/images is taken from the tag itself, so the
`version/VERSION` file does not strictly need to be bumped first.

## Running the container

```sh
docker run --rm -p 8500:8500 ghcr.io/opensoba/opensoba:latest
```

The image runs as a non-root `soba` user, stores state in `/soba/data`
(mount a volume there) and reads extra config from `/soba/config`.
`SOBA_*` environment variables are supported, with the legacy `CONSUL_*`
spelling still accepted for compatibility.

## Manual / no-network builds

You don't need the workflow to build locally:

```sh
make dev          # -> ./bin/soba
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o soba .
docker build -t opensoba .
```
