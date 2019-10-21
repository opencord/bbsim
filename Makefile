# Copyright 2019-present Open Networking Foundation
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

VERSION     ?= $(shell cat ./VERSION)
DIFF		?= $(git diff --shortstat 2> /dev/null | tail -n1)
GIT_STATUS	?= $(shell [ -z "$DIFF" ] && echo "Dirty" || echo "Clean")

## Docker related
DOCKER_TAG  			?= ${VERSION}
DOCKER_REPOSITORY  		?= ""
DOCKER_REGISTRY 		?= ""
DOCKER_RUN_ARGS			?= ""
DOCKER_PORTS			?= -p 50070:50070 -p 50060:50060 -p 50071:50071 -p 50072:50072 -p 50073:50073

## protobuf related
VOLTHA_PROTOS			?= $(shell GO111MODULE=on go list -f '{{ .Dir }}' -m github.com/opencord/voltha-protos)
GOOGLEAPI				?= $(shell GO111MODULE=on go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)
TOOLS_DIR := tools
TOOLS_BIN := $(TOOLS_DIR)/bin/

export PATH=$(shell echo $$PATH):$(PWD)/$(TOOLS_BIN)

# Public targets
all: help

# go installed tools.go
GO_TOOLS := github.com/golang/protobuf/protoc-gen-go \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
            github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# tools
GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN), $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))

TEST_PACKAGES := github.com/opencord/bbsim/cmd/... \
                 github.com/opencord/bbsim/internal/...

setup_tools: $(GO_TOOLS_BIN)

$(GO_TOOLS_BIN): $(GO_TOOLS_VENDOR)
	GO111MODULE=on GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

protos: dep setup_tools api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go # @HELP Build proto files

dep: # @HELP Download the dependencies to the vendor folder
	GO111MODULE=on go mod vendor

_build: dep protos fmt build-bbsim build-bbsimctl build-bbr

.PHONY: build
build: # @HELP Build the binaries (it runs inside a docker container and output the built code on your local file system)
	docker build -t ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim-builder:${DOCKER_TAG} -f build/ci/Dockerfile.builder .
	docker run --rm -v $(shell pwd):/bbsim ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim-builder:${DOCKER_TAG} /bin/sh -c "cd /bbsim; make _build"

test: dep protos fmt # @HELP Execute unit tests
	GO111MODULE=on go test -v -mod vendor $(TEST_PACKAGES) -covermode count -coverprofile ./tests/results/go-test-coverage.out 2>&1 | tee ./tests/results/go-test-results.out
	go-junit-report < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml
	gocover-cobertura < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml

fmt:
	go fmt ./...

docker-build: # @HELP Build the BBSim docker container (contains BBSimCtl too)
	docker build -t ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} -f build/package/Dockerfile .

docker-push: # @HELP Push the docker container to a registry
	docker push ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG}

docker-run: # @HELP Runs the container locally (available options: DOCKER_RUN_ARGS="-pon 2 -onu 2" make docker-run)
	docker run -d ${DOCKER_PORTS} --privileged --rm --name bbsim ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} /app/bbsim ${DOCKER_RUN_ARGS}

docker-run-dev: # @HELP Runs the container locally (intended for development purposes, not in detached mode)
	docker run ${DOCKER_PORTS} --privileged --rm --name bbsim ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} /app/bbsim ${DOCKER_RUN_ARGS}

.PHONY: docs
docs: swagger # @HELP Generate docs and opens them in the browser
	cd docs; make doc_venv; make html; cd -
	@echo -e "\nBBSim documentation generated in file://${PWD}/docs/build/html/index.html"

# Release related items
# Generates binaries in $RELEASE_DIR with name $RELEASE_NAME-$RELEASE_OS_ARCH
# Inspired by: https://github.com/kubernetes/minikube/releases
RELEASE_DIR     ?= release
RELEASE_OS_ARCH ?= linux-amd64 linux-arm64 windows-amd64 darwin-amd64

RELEASE_BBR_NAME    ?= bbr
RELEASE_BBR_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(RELEASE_BBR_NAME)-$(rel))
RELEASE_BBSIM_NAME    ?= bbsimctl
RELEASE_BBSIM_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(RELEASE_BBSIM_NAME)-$(rel))

$(RELEASE_BBR_BINS):
	export GOOS=$(rel_os) ;\
	export GOARCH=$(rel_arch) ;\
	GO111MODULE=on go build -i -v -mod vendor \
        	-ldflags "-w -X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
        		-X main.gitStatus=${GIT_STATUS} \
        		-X main.version=${VERSION}" \
        	-o "$@" ./cmd/bbr

$(RELEASE_BBSIM_BINS):
	export GOOS=$(rel_os) ;\
	export GOARCH=$(rel_arch) ;\
	GO111MODULE=on go build -i -v -mod vendor \
        	-ldflags "-w -X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
        		-X main.gitStatus=${GIT_STATUS} \
        		-X main.version=${VERSION}" \
        	-o "$@" ./cmd/bbsim

.PHONY: release $(RELEASE_BBR_BINS) $(RELEASE_BBSIM_BINS)
release: dep protos $(RELEASE_BBR_BINS) $(RELEASE_BBSIM_BINS) # @HELP Creates release ready bynaries for BBSimctl and BBR artifacts
swagger: docs/swagger/bbsim/bbsim.swagger.json docs/swagger/leagacy/bbsim.swagger.json # @HELP Generate swagger documentation for BBSim API

help: # @HELP Print the command options
	@echo
	@echo "\033[0;31m    BroadBand Simulator (BBSim) \033[0m"
	@echo
	@echo Emulates the control plane of an openolt compatible device
	@echo Useful for development and scale testing
	@echo
	@grep -E '^.*: .* *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": .* *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '


# Internals

build-bbr:
	GO111MODULE=on go build -i -v -mod vendor \
    	-ldflags "-w -X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
    		-X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
    		-X main.gitStatus=${GIT_STATUS} \
    		-X main.version=${VERSION}" \
    	./cmd/bbr

build-bbsim:
	GO111MODULE=on go build -i -v -mod vendor \
    	-ldflags "-w -X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
    		-X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
    		-X main.gitStatus=${GIT_STATUS} \
    		-X main.version=${VERSION}" \
    	./cmd/bbsim

build-bbsimctl:
	GO111MODULE=on go build -i -v -mod vendor \
		-ldflags "-w -X github.com/opencord/bbsim/internal/bbsimctl/config.BuildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
			-X github.com/opencord/bbsim/internal/bbsimctl/config.CommitHash=$(shell git log --pretty=format:%H -n 1) \
			-X github.com/opencord/bbsim/internal/bbsimctl/config.GitStatus=${GIT_STATUS} \
			-X github.com/opencord/bbsim/internal/bbsimctl/config.Version=${VERSION}" \
		./cmd/bbsimctl

api/openolt/openolt.pb.go: api/openolt/openolt.proto
	@protoc -I. \
    	-I${GOOGLEAPI}/third_party/googleapis \
    	--go_out=plugins=grpc:./ \
    	$<

api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go: api/bbsim/bbsim.proto api/bbsim/bbsim.yaml
	@protoc -I. \
		-I${GOOGLEAPI}/third_party/googleapis \
    	--go_out=plugins=grpc:./ \
		--grpc-gateway_out=logtostderr=true,grpc_api_configuration=api/bbsim/bbsim.yaml,allow_delete_body=true:./ \
    	$<

api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go: api/legacy/bbsim.proto
	@protoc -I. \
		-I${GOOGLEAPI}/third_party/googleapis/ \
		-I${GOOGLEAPI}/ \
		-I${VOLTHA_PROTOS}/protos/ \
    	--go_out=plugins=grpc:./ \
		--grpc-gateway_out=logtostderr=true,allow_delete_body=true:./ \
    	$<

docs/swagger/bbsim/bbsim.swagger.json: api/bbsim/bbsim.yaml
	@protoc -I ./api \
	-I${GOOGLEAPI}/ \
	-I${VOLTHA_PROTOS}/protos/ \
	--swagger_out=logtostderr=true,allow_delete_body=true,grpc_api_configuration=$<:docs/swagger/ \
	api/bbsim/bbsim.proto

docs/swagger/leagacy/bbsim.swagger.json: api/legacy/bbsim.proto
	@protoc -I ./api \
	-I${GOOGLEAPI}/ \
	-I${VOLTHA_PROTOS}/protos/ \
	--swagger_out=logtostderr=true,allow_delete_body=true:docs/swagger/ \
	$<