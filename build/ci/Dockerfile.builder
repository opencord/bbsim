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

# this container is used to provide a build environment so that we don't need
# the prereqs installed on a CI machine

FROM golang:1.12-stretch

# install prereqs
ENV PROTOC_VERSION 3.6.1
ENV PROTOC_SHA256SUM 6003de742ea3fcf703cfec1cd4a3380fd143081a2eb0e559065563496af27807

RUN apt-get update \
 && apt-get install -y unzip libpcap-dev \
 && curl -L -o /tmp/protoc-${PROTOC_VERSION}-linux-x86_64.zip https://github.com/google/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip \
 && echo "$PROTOC_SHA256SUM  /tmp/protoc-${PROTOC_VERSION}-linux-x86_64.zip" | sha256sum -c - \
 && unzip /tmp/protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /tmp/protoc3 \
 && mv /tmp/protoc3/bin/* /usr/local/bin/ \
 && mv /tmp/protoc3/include/* /usr/local/include/ \
 && go get -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
 && go get -v github.com/golang/protobuf/protoc-gen-go