# Copyright 2018-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# bbsim dockerfile

# builder parent
FROM golang:1.13.8-stretch as builder

RUN apt-get update && apt-get install -y build-essential=12.3 libpcap-dev=1.8.1-3 --no-install-recommends

ENV CGO_ENABLED=1

WORKDIR /app

# copy and build
COPY . ./

RUN go build -mod vendor \
      -ldflags "-w -X main.buildTime=$(date +%Y/%m/%d-%H:%M:%S) \
        -X main.commitHash=$(git log --pretty=format:%H -n 1) \
        -X main.gitStatus=${GIT_STATUS} \
        -X main.version=${VERSION}" \
      ./cmd/bbsim && \
    go build -mod vendor \
      -ldflags "-w -X github.com/opencord/bbsim/internal/bbsimctl/config.BuildTime=$(date +%Y/%m/%d-%H:%M:%S) \
        -X github.com/opencord/bbsim/internal/bbsimctl/config.CommitHash=$(git log --pretty=format:%H -n 1) \
        -X github.com/opencord/bbsim/internal/bbsimctl/config.GitStatus=${GIT_STATUS} \
        -X github.com/opencord/bbsim/internal/bbsimctl/config.Version=${VERSION}" \
      ./cmd/bbsimctl

# runtime parent
FROM ubuntu:18.04

# runtime prereqs
# the symlink on libpcap is because both alpine and debian come with 1.8.x, but
# debian symlinks it to 0.8 for historical reasons:
# https://packages.debian.org/stretch/libpcap0.8-dev
RUN apt-get update \
 && apt-get install -y libpcap-dev=1.8.1-6ubuntu1.18.04.1 isc-dhcp-server=4.3.5-3ubuntu7.1 \
 network-manager=1.10.6-2ubuntu1.4 tcpdump=4.9.3-0ubuntu0.18.04.1 --no-install-recommends \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* \
 && ln -s /usr/lib/libpcap.so.1.8.1 /usr/lib/libpcap.so.0.9

COPY ./configs/dhcpd.conf /etc/dhcp/
RUN mv /usr/sbin/dhcpd /usr/local/bin/ \
&& mv /sbin/dhclient /usr/local/bin/ \
&& touch /var/lib/dhcp/dhcpd.leases

WORKDIR /app
COPY --from=builder /app/bbsim /app/bbsim
COPY --from=builder /app/bbsimctl /usr/bin/bbsimctl
RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump
RUN chmod a+x /app/bbsim
RUN chmod a+x /usr/bin/bbsimctl
COPY ./configs/ ./configs/
RUN bbsimctl completion bash >> "$HOME/.bashrc"
CMD [ "/app/bbsim" ]
