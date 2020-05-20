# Copyright 2020-present Open Networking Foundation
# SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

ARG VOLTHA_TOOLS_VERSION
FROM voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint

# install repo-specific C dependencies
RUN apk add --no-cache build-base libpcap-dev
