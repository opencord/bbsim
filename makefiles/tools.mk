# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2023 Open Networking Foundation (ONF) and the ONF Contributors
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
# limitations under the License.d
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

## Docker related -- are these ${shell vars} or $(make macros) ?
DOCKER_TAG  			?= ${VERSION}
DOCKER_REPOSITORY  		?=
DOCKER_REGISTRY 		?=
DOCKER_RUN_ARGS			?=
DOCKER_PORTS			?= -p 50070:50070 -p 50060:50060 -p 50071:50071 -p 50072:50072 -p 50073:50073 -p 50074:50074 -p 50075:50075
TYPE                            ?= minimal

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.3.1

docker-run     = docker run --rm --user $$(id -u):$$(id -g)#   # Docker command stem
docker-run-app = $(docker-run) -v ${CURDIR}:/app#              # w/filesystem mount
is-stdin       = $(shell test -t 0 && echo "-it")

# Docker volume mounts: container:/app/release <=> localhost:{pwd}/release
vee-golang     = -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg
vee-citools    = voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}

GO                = $(docker-run-app) $(is-stdin) -v gocache:/.cache $(vee-golang) $(vee-citools)-golang go
GO_SH             = $(docker-run-app) $(is-stdin) --env-file ./foobar -v gocache:/.cache $(vee-golang) $(vee-citools)-golang sh -c
GO_JUNIT_REPORT   = $(docker-run-app) -i $(vee-citools)-go-junit-report go-junit-report
GOCOVER_COBERTURA = $(docker-run) -v ${CURDIR}:/app/src/github.com/opencord/bbsim -i $(vee-citools)-gocover-cobertura gocover-cobertura
GOLANGCI_LINT     = $(docker-run-app) $(is-stdin) -v gocache:/.cache $(vee-golang) $(vee-citools)-golangci-lint golangci-lint
PROTOC            = $(docker-run-app) $(is-stdin) $(vee-golang) $(vee-citools)-protoc protoc

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
