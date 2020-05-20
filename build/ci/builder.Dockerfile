# Copyright 2020-present Open Networking Foundation
# SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.13.8-stretch

RUN apt-get update && apt-get install -y \
    build-essential libpcap-dev && \
    mkdir -m 777 /.cache /go/pkg

ENV CGO_ENABLED=1

WORKDIR /app