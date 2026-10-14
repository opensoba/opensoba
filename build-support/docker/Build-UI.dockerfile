# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Official Node.js 22 LTS image. It ships yarn (classic), make and git, which
# are all that `make dist-docker` needs. Browsers are not required: this image
# is only used to build the UI (see build_ui in build-support/functions), not to
# run the browser-based test suite.
FROM node:22-bookworm

USER root

RUN mkdir /consul-src
WORKDIR /consul-src
CMD make dist-docker
