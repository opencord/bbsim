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
GIT_STATUS	?= $(shell [[ $DIFF != "" ]] && echo "Dirty" || echo "Clean")

## Docker related
DOCKER_TAG  			?= ${VERSION}
DOCKER_REPOSITORY  		?= ""
DOCKER_REGISTRY 		?= ""
DOCKER_RUN_ARGS			?= ""

# Public targets

all: help

protos: api/bbsim/bbsim.pb.go # @HELP Build proto files

dep: # @HELP Download the dependencies to the vendor folder
	GO111MODULE=on go mod vendor

build: dep protos fmt build-bbsim build-bbsimctl build-bbr # @HELP Build the binary

test: dep protos fmt # @HELP Execute unit tests
	GO111MODULE=on go test -v -mod vendor ./... -covermode count -coverprofile ./tests/results/go-test-coverage.out 2>&1 | tee ./tests/results/go-test-results.out
	go-junit-report < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml
	gocover-cobertura < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml

fmt:
	go fmt ./...

docker-build: # @HELP Build a docker container
	docker build -t ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} -f build/package/Dockerfile .

docker-push: # @HELP Push a docker container to a registry
	docker push ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG}

docker-run: # @HELP Run the container locally (intended for development purposes: DOCKER_RUN_ARGS="-pon 2 -onu 2" make docker-run)
	docker run -d -p 50070:50070 -p 50060:50060 --privileged --rm --name bbsim ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} /app/bbsim ${DOCKER_RUN_ARGS}

.PHONY: docs
docs: # @HELP Generate docs and serve them in the browser
	@echo "\033[36m Open your browser at localhost:6060 \033[0m"
	godoc -http=localhost:6060

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
        		-X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
        		-X main.gitStatus=${GIT_STATUS} \
        		-X main.version=${VERSION}" \
        	-o "$@" ./cmd/bbr

$(RELEASE_BBSIM_BINS):
	export GOOS=$(rel_os) ;\
	export GOARCH=$(rel_arch) ;\
	GO111MODULE=on go build -i -v -mod vendor \
        	-ldflags "-w -X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
        		-X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
        		-X main.gitStatus=${GIT_STATUS} \
        		-X main.version=${VERSION}" \
        	-o "$@" ./cmd/bbsim

.PHONY: release $(RELEASE_BBR_BINS) $(RELEASE_BBSIM_BINS)
release: $(RELEASE_BBR_BINS) $(RELEASE_BBSIM_BINS) # @HELP Release BBSimctl and BBR artifacts

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
	@protoc -I . \
    	-I${GOPATH}/src \
    	-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    	--go_out=plugins=grpc:./ \
    	$<

api/bbsim/bbsim.pb.go: api/bbsim/bbsim.proto
	@protoc -I . \
    	-I${GOPATH}/src \
    	-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    	--go_out=plugins=grpc:./ \
    	$<

