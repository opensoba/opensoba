# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

FROM circleci/node:16-browsers

USER root

RUN mkdir /consul-src
WORKDIR /consul-src
CMD make dist-docker
