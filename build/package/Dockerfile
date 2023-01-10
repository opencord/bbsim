# Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors
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
WORKDIR /app
COPY --from=builder /app/bbsim /app/bbsim
COPY --from=builder /app/bbsimctl /usr/bin/bbsimctl
COPY ./configs/ ./configs/
RUN chmod a+x /app/bbsim
RUN chmod a+x /usr/bin/bbsimctl
RUN bbsimctl completion bash >> "$HOME/.bashrc"
CMD [ "/app/bbsim" ]

# Label image
ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_url=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_label_schema_build_date=unknown
ARG org_opencord_vcs_commit_date=unknown
ARG org_opencord_vcs_dirty=unknown

LABEL org.label-schema.schema-version=1.0 \
      org.label-schema.name=voltha-rw-core \
      org.label-schema.version=$org_label_schema_version \
      org.label-schema.vcs-url=$org_label_schema_vcs_url \
      org.label-schema.vcs-ref=$org_label_schema_vcs_ref \
      org.label-schema.build-date=$org_label_schema_build_date \
      org.opencord.vcs-commit-date=$org_opencord_vcs_commit_date \
      org.opencord.vcs-dirty=$org_opencord_vcs_dirty
