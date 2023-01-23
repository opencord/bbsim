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
# limitations under the License.
# -----------------------------------------------------------------------

.DEFAULT_GOAL   := help
MAKECMDGOALS    ?= help

TOP     ?= .
MAKEDIR ?= $(TOP)/makefiles

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
help-targets := help-all help HELP
$(if $(filter $(help-targets),$(MAKECMDGOALS))\
    ,$(eval include $(MAKEDIR)/help.mk))

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
SHELL           = bash -e -o pipefail
VERSION         ?= $(shell cat ./VERSION)
DIFF		?= $(git diff --shortstat 2> /dev/null | tail -n1)
GIT_STATUS	?= $(shell [ -z "$DIFF" ] && echo "Dirty" || echo "Clean")

## Docker related -- are these ${shell vars} or $(make macros) ?
DOCKER_TAG  			?= ${VERSION}
DOCKER_REPOSITORY  		?=
DOCKER_REGISTRY 		?=
DOCKER_RUN_ARGS			?=
DOCKER_PORTS			?= -p 50070:50070 -p 50060:50060 -p 50071:50071 -p 50072:50072 -p 50073:50073 -p 50074:50074 -p 50075:50075
TYPE                            ?= minimal

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.3.1

GO                = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_SH             = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang sh -c '# fix-editor-colorization-quote(')
GO_JUNIT_REPORT   = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app/src/github.com/opencord/bbsim -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOLANGCI_LINT     = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint golangci-lint
HADOLINT          = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-hadolint hadolint
PROTOC            = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-protoc protoc

## use local vars to shorten paths
bbsim-tag = ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG}

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------

# Public targets
all: help

protos: api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go api/bbsim/bbsim_dmi.pb.go # @HELP Build proto files

.PHONY: build
build: protos build-bbsim build-bbsimctl build-bbr

## lint and unit tests

lint-dockerfile:
	@echo "Running Dockerfile lint check..."
	@${HADOLINT} $$(find ./build -name "Dockerfile*")
	@echo "Dockerfile lint check OK"

lint-mod:
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)
	${GO} mod tidy
	${GO} mod vendor
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."

lint: lint-mod lint-dockerfile

sca:
	@$(RM) -r ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run -vv --deadline=6m --out-format junit-xml ./... | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

test: docs-lint test-unit test-bbr

test-unit: clean local-omci-lib-go # @HELP Execute unit tests
	@echo "Running unit tests..."
	@mkdir -p ./tests/results
	@${GO} test -mod=vendor -bench=. -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

test-bbr: release-bbr docker-build # @HELP Validate that BBSim and BBR are working together
	DOCKER_RUN_ARGS="-pon 2 -onu 2" $(MAKE) docker-run
	sleep 5
	./$(RELEASE_DIR)/$(RELEASE_BBR_NAME)-linux-amd64 -pon 2 -onu 2 -logfile tmp.logs
	docker $(RM) -f bbsim

mod-update: # @HELP Download the dependencies to the vendor folder
	${GO} mod tidy
	${GO} mod vendor

docker-build: local-omci-lib-go local-protos# @HELP Build the BBSim docker container (contains BBSimCtl too)
	docker build \
	  -t "$(bbsim-tag)" \
	  -f build/package/Dockerfile .

docker-push: # @HELP Push the docker container to a registry
	docker push "$(bbsim-tag)"
docker-kind-load:
	@if [ "`kind get clusters | grep voltha-$(TYPE)`" = '' ]; then echo "no voltha-$(TYPE) cluster found" && exit 1; fi
	kind load docker-image "$(bbsim-tag)" --name=voltha-$(TYPE) --nodes $(shell kubectl get nodes --template='{{range .items}}{{.metadata.name}},{{end}}' | sed 's/,$$//')

## -----------------------------------------------------------------------
## docker-run
## -----------------------------------------------------------------------

# % docker run --detach unless dev-mode
is-docker-run := $(filter-out %-dev,$(MAKECMDGOALS))
is-docker-run := $(findstring docker-run,$(is-docker-run))

docker-run-cmd  = docker run
$(if $(is-docker-run),$(eval docker-run-cmd += --detach))
docker-run-cmd += ${DOCKER_PORTS}
docker-run-cmd += --privileged
docker-run-cmd += --rm
docker-run-cmd += --name bbsim
docker-run-cmd += "$(bbsim-tag)" /app/bbsim
docker-run-cmd += ${DOCKER_RUN_ARGS}

docker-run: # @HELP Runs the container locally (available options: DOCKER_RUN_ARGS="-pon 2 -onu 2" make docker-run)
	docker ps
	$(docker-run-cmd)

docker-run-dev: # @HELP Runs the container locally (intended for development purposes, not in detached mode)
	$(docker-run-cmd)

## -----------------------------------------------------------------------
## docker-run
## -----------------------------------------------------------------------
.PHONY: docs docs-lint
docs: swagger # @HELP Generate docs and opens them in the browser
	$(MAKE) -C docs html
	@echo -e "\nBBSim documentation generated in file://${PWD}/docs/build/html/index.html"

docs-lint:
	$(MAKE) -C docs lint

# Release related items
# Generates binaries in $RELEASE_DIR with name $RELEASE_NAME-$RELEASE_OS_ARCH
# Inspired by: https://github.com/kubernetes/minikube/releases
RELEASE_DIR     ?= release
RELEASE_OS_ARCH ?= linux-amd64 linux-arm64 windows-amd64 darwin-amd64

RELEASE_BBR_NAME      ?= bbr
RELEASE_BBSIM_NAME    ?= bbsim
RELEASE_BBSIMCTL_NAME ?= bbsimctl

release-bbr:
	@echo "$(RELEASE_BBR_NAME)-linux-amd64"
	${GO} build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  -o "$(RELEASE_DIR)/$(RELEASE_BBR_NAME)-linux-amd64" ./cmd/bbr

release-bbsim:
	@echo "$(RELEASE_BBSIM_NAME)-linux-amd64"
	${GO} build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  -o "$(RELEASE_DIR)/$(RELEASE_BBSIM_NAME)-linux-amd64" ./cmd/bbsim

release-bbsimctl:	
	@echo "** $(MAKE): processing target [$@]"
	${GO_SH} set -eo pipefail; \
  for os_arch in ${RELEASE_OS_ARCH}; do \
	    echo "$(RELEASE_BBSIMCTL_NAME)-$$os_arch"; \
	    GOOS="$${os_arch%-*}" GOARCH="$${os_arch#*-}" go build -mod vendor \
	      -ldflags "-w -X github.com/opencord/bbsim/internal/bbsimctl/config.BuildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	      -X github.com/opencord/bbsim/internal/bbsimctl/config.CommitHash=$(shell git log --pretty=format:%H -n 1) \
	      -X github.com/opencord/bbsim/internal/bbsimctl/config.GitStatus=${GIT_STATUS} \
	      -X github.com/opencord/bbsim/internal/bbsimctl/config.Version=${VERSION}" \
	    -o "$(RELEASE_DIR)/$(RELEASE_BBSIMCTL_NAME)-$$os_arch" ./cmd/bbsimctl; \
	  done'
# fix-editor-colorization-quote(')

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
release-deps += release-bbr
release-deps += release-bbsim
release-deps += release-bbsimctl
.PHONY: release $(release-deps)
release: release-init $(release-deps) # @HELP Creates release ready bynaries for BBSimctl and BBR artifacts

$(release-deps) : release-init

release-init:
	@echo "** $(MAKE): processing target [$@]"
	${GO_SH} set -eo pipefail\
; echo "PWD: $$(/bin/pwd)"\
; mkdir -p $(RELEASE_DIR)'
# fix-editor-colorization-quote(')

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
swagger-deps += docs/swagger/bbsim/bbsim.swagger.json
swagger-deps += docs/swagger/leagacy/bbsim.swagger.json
.PHONY: $(swagger-deps)
swagger: $(swagger-deps) # @HELP Generate swagger documentation for BBSim API

## -----------------------------------------------------------------------
## Local Development Helpers
## -----------------------------------------------------------------------
local-omci-lib-go:
ifdef LOCAL_OMCI_LIB_GO
	mkdir -p vendor/github.com/opencord/omci-lib-go
	cp -r ${LOCAL_OMCI_LIB_GO}/* vendor/github.com/opencord/omci-lib-go
endif

local-protos: ## Copies a local version of the voltha-protos dependency into the vendor directory
ifdef LOCAL_PROTOS
	$(RM) -r vendor/github.com/opencord/voltha-protos/v5/go
	mkdir -p vendor/github.com/opencord/voltha-protos/v5/go
	cp -r ${LOCAL_PROTOS}/go/* vendor/github.com/opencord/voltha-protos/v5/go
	$(RM) -r vendor/github.com/opencord/voltha-protos/v5/go/vendor
endif

# Internals

clean:
	@$(RM) -f bbsim
	@$(RM) -f bbsimctl
	@$(RM) -f bbr
	@$(RM) -r tools/bin
	@$(RM) -r release/*

build-bbr: local-omci-lib-go local-protos
	@go build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  ./cmd/bbr

build-bbsim:
	@go build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  ./cmd/bbsim

build-bbsimctl:
	@go build -mod vendor \
	  -ldflags "-w -X github.com/opencord/bbsim/internal/bbsimctl/config.BuildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.CommitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.GitStatus=${GIT_STATUS} \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.Version=${VERSION}" \
	  ./cmd/bbsimctl

setup_tools:
	@echo "Downloading dependencies..."
	@${GO} mod download github.com/grpc-ecosystem/grpc-gateway github.com/opencord/voltha-protos/v5
	@echo "Dependencies downloaded OK"

VOLTHA_PROTOS ?= $(shell ${GO} list -f '{{ .Dir }}' -m github.com/opencord/voltha-protos/v5)
GOOGLEAPI     ?= $(shell ${GO} list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)

.PHONY: api/openolt/openolt.pb.go api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go
api/openolt/openolt.pb.go: api/openolt/openolt.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
      -I${GOOGLEAPI}/third_party/googleapis \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
      $<

api/bbsim/bbsim_dmi.pb.go: api/bbsim/bbsim_dmi.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
      -I${GOOGLEAPI}/third_party/googleapis \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
      $<

api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go: api/bbsim/bbsim.proto api/bbsim/bbsim.yaml setup_tools
	@echo $@
	@${PROTOC} -I. \
	  -I${GOOGLEAPI}/third_party/googleapis \
	  -I${VOLTHA_PROTOS}/protos/ \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
	  --grpc-gateway_out=logtostderr=true,paths=source_relative,grpc_api_configuration=api/bbsim/bbsim.yaml,allow_delete_body=true:./ \
      $<

api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go: api/legacy/bbsim.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
	  -I${GOOGLEAPI}/third_party/googleapis/ \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
      --go_out=plugins=grpc:./ \
	  --grpc-gateway_out=logtostderr=true,paths=source_relative,allow_delete_body=true:./ \
      $<

docs/swagger/bbsim/bbsim.swagger.json: api/bbsim/bbsim.yaml setup_tools
	@echo $@
	@${PROTOC} -I ./api \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
	  --swagger_out=logtostderr=true,allow_delete_body=true,disable_default_errors=true,grpc_api_configuration=$<:docs/swagger/ \
	  api/bbsim/bbsim.proto

docs/swagger/leagacy/bbsim.swagger.json: api/legacy/bbsim.proto setup_tools
	@echo $@
	@${PROTOC} -I ./api \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
	  --swagger_out=logtostderr=true,allow_delete_body=true,disable_default_errors=true:docs/swagger/ \
	  $<

# [EOF]
