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
DOCKER_REPOSITORY  		?= voltha/
DOCKER_REGISTRY 		?= ""

# Public targets

all: help

protos: api/bbsim/bbsim.pb.go # @HELP Build proto files

build: protos # @HELP Build the binary
	GO111MODULE=on go build -i -v \
	-ldflags "-X main.buildTime=$(shell date +”%Y/%m/%d-%H:%M:%S”) \
		-X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
		-X main.gitStatus=${GIT_STATUS} \
		-X main.version=${VERSION}" \
	-o ./cmd/bbsim ./internal/bbsim

test: protos # @HELP Execute unit tests
	GO111MODULE=on go test ./internal/bbsim

docker-build: # @HELP Build a docker container
	docker build --build-arg GIT_STATUS=${GIT_STATUS} -t ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG} -f build/package/Dockerfile .

docker-push: # @HELP Push a docker container to a registry
	docker push ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG}

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

