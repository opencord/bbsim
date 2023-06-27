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

$(if $(DEBUG),$(warning ENTER))

.DEFAULT_GOAL   := help
MAKECMDGOALS    ?= help

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
TOP ?=$(strip \
  $(dir \
    $(abspath $(lastword $(MAKEFILE_LIST)))\
   )\
)

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(TOP)/config.mk
include $(TOP)/makefiles/include.mk
include $(ONF_MAKEDIR)/release/include.mk

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
help-targets := help-all help HELP
$(if $(filter $(help-targets),$(MAKECMDGOALS))\
    ,$(eval include $(MAKEDIR)/help.mk))

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
VERSION         ?= $(shell cat ./VERSION)
DIFF		?= $(git diff --shortstat 2> /dev/null | tail -n1)
GIT_STATUS	?= $(shell [ -z "$DIFF" ] && echo "Dirty" || echo "Clean")

include $(MAKEDIR)/tools.mk   # Command macros for DOCKER_*, GO_*

## use local vars to shorten paths
bbsim-tag = ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}bbsim:${DOCKER_TAG}

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------

# Public targets
all: help

protos: api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go api/bbsim/bbsim_dmi.pb.go # @HELP Build proto files

## -----------------------------------------------------------------------
## Intent: Build tools for use on localhost
## -----------------------------------------------------------------------
build-target-deps += build-bbsim
build-target-deps += build-bbsimctl
build-target-deps += build-bbr

.PHONY: build $(build-target-deps)
build: protos $(build-target-deps)

## -----------------------------------------------------------------------
## Intent: Cross-compile binaries for release
## -----------------------------------------------------------------------
release-target-deps += release-bbr
release-target-deps += release-bbsim
release-target-deps += release-bbsimctl

.PHONY: release $(release-target-deps)
release: $(release-target-deps)

## lint and unit tests

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint-dockerfile:
	@echo "Running Dockerfile lint check..."
	@${HADOLINT} $$(find ./build -name "Dockerfile*")
	@echo "Dockerfile lint check OK"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint-mod:
	$(call banner-enter,Target $@)
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)

	$(HIDE)$(MAKE) --no-print-directory mod-update

	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."
	$(call banner-enter,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint: lint-mod lint-dockerfile

## -----------------------------------------------------------------------
## Intent: Generate a static code analysis report
## -----------------------------------------------------------------------
sca-report-dir := ./sca-report
sca-report-xml := $(sca-report-dir)/sca-report.xml
sca:
	$(call begin-enter,Target $@)
	@$(RM) -r $(sca-report-dir)
	@mkdir -p $(sca-report-dir)
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run -vv --deadline=6m --out-format junit-xml ./... \
	  | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"
	$(call begin-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
test: docs-lint test-unit test-bbr

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
coverage-out := ./tests/results/go-test-coverage.out
coverage-xml := ./tests/results/go-test-coverage.xml

results-out  := ./tests/results/go-test-results.out
results-xml  := ./tests/results/go-test-results.xml

test-unit: clean local-omci-lib-go # @HELP Execute unit tests

	$(call banner-enter,Target $@)

	@mkdir -p $(dir $(results-out))

	$(call banner,Running unit tests...)

	$(if $(LOCAL_FIX_PERMS),\
	    chmod o+w $(dir $(results-out) $(results-out)))
	set -euo pipefail \
	&& ${GO} test -mod=vendor -bench=. -v \
	  -coverprofile $(coverage-out) \
	  -covermode count ./... \
	  2>&1 | tee "$(results-out)"
	$(if $(LOCAL_FIX_PERMS),\
	    chmod o-w $(dir $(results-out) $(results-out)))

	$(call banner,Coverage report: junit)
	$(if $(LOCAL_FIX_PERMS),\
	    chmod o+w $(dir $(results-xml) $(results-xml)))
	${GO_JUNIT_REPORT} < $(results-out) > $(results-xml)
	$(if $(LOCAL_FIX_PERMS),\
	    chmod o-w $(dir $(results-xml) $(results-xml)))

	$(call banner,Coverage report: Cobertura)
	$(if $(LOCAL_FIX_PERMS),\
	    chmod o+w $(dir $(results-xml) $(coverage-xml)))
	${GOCOVER_COBERTURA} < $(coverage-out) > $(coverage-xml)
	$(if $(LOCAL_FIX_PERMS),\
	    chmod o-w $(dir $(results-xml) $(coverage-xml)))	

	$(call banner,Coverage report directory perms:)

	@echo
	@/bin/ls -ld tests tests/results

	@echo
	@find $(dir $(results-out)) -print0 | xargs -0 ls -l
        # chown -R $$(id -u):$$(id -g) $(dir $(results-xml)

	$(call banner,Locally modified files: git status)
	git status

	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
test-bbr: release-bbr docker-build # @HELP Validate that BBSim and BBR are working together
	DOCKER_RUN_ARGS="-pon 2 -onu 2" $(MAKE) docker-run
	sleep 5
	./$(RELEASE_DIR)/$(RELEASE_BBR_NAME)-linux-amd64 -pon 2 -onu 2 -logfile tmp.logs
	docker $(RM) -f bbsim

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docker-build: local-omci-lib-go local-protos# @HELP Build the BBSim docker container (contains BBSimCtl too)
	docker build \
	  -t "$(bbsim-tag)" \
	  -f build/package/Dockerfile .

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docker-push: # @HELP Push the docker container to a registry
	docker push "$(bbsim-tag)"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
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

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docker-run: # @HELP Runs the container locally (available options: DOCKER_RUN_ARGS="-pon 2 -onu 2" make docker-run)
	docker ps
	$(docker-run-cmd)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docker-run-dev: # @HELP Runs the container locally (intended for development purposes, not in detached mode)
	$(docker-run-cmd)

## -----------------------------------------------------------------------
## docker-run
## -----------------------------------------------------------------------
.PHONY: docs docs-lint
docs: swagger # @HELP Generate docs and opens them in the browser

	$(call begin-enter,Target $@)
	-$(MAKE) -C docs html
	$(call begin-leave,Target $@)

	@echo -e "\nBBSim documentation generated in file://${PWD}/docs/build/html/index.html"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docs-lint:
	$(MAKE) -C docs lint-doc8

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

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
local-protos: ## Copies a local version of the voltha-protos dependency into the vendor directory
ifdef LOCAL_PROTOS
	$(RM) -r vendor/github.com/opencord/voltha-protos/v5/go
	mkdir -p vendor/github.com/opencord/voltha-protos/v5/go
	cp -r ${LOCAL_PROTOS}/go/* vendor/github.com/opencord/voltha-protos/v5/go
	$(RM) -r vendor/github.com/opencord/voltha-protos/v5/go/vendor
endif

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
clean ::
	@$(RM) -f bbsim
	@$(RM) -f bbsimctl
	@$(RM) -f bbr
	@$(RM) -r tools/bin
	@$(RM) -r release

	@$(RM) $(coverage-out)
	@$(RM) $(coverage-xml)

	@$(RM) $(results-out)
	@$(RM) $(results-xml)

	@$(RM) $(sca-report-xml)

sterile :: clean

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
setup_tools:
	@echo "Downloading dependencies..."
	@${GO} mod download github.com/grpc-ecosystem/grpc-gateway github.com/opencord/voltha-protos/v5
	@echo "Dependencies downloaded OK"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
VOLTHA_PROTOS ?= $(shell ${GO} list -f '{{ .Dir }}' -m github.com/opencord/voltha-protos/v5)
GOOGLEAPI     ?= $(shell ${GO} list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)

.PHONY: api/openolt/openolt.pb.go api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go
api/openolt/openolt.pb.go: api/openolt/openolt.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
      -I${GOOGLEAPI}/third_party/googleapis \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
      $<

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
api/bbsim/bbsim_dmi.pb.go: api/bbsim/bbsim_dmi.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
      -I${GOOGLEAPI}/third_party/googleapis \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
      $<

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
api/bbsim/bbsim.pb.go api/bbsim/bbsim.pb.gw.go: api/bbsim/bbsim.proto api/bbsim/bbsim.yaml setup_tools
	@echo $@
	@${PROTOC} -I. \
	  -I${GOOGLEAPI}/third_party/googleapis \
	  -I${VOLTHA_PROTOS}/protos/ \
      --go_out=plugins=grpc:./ --go_opt=paths=source_relative \
	  --grpc-gateway_out=logtostderr=true,paths=source_relative,grpc_api_configuration=api/bbsim/bbsim.yaml,allow_delete_body=true:./ \
      $<

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
api/legacy/bbsim.pb.go api/legacy/bbsim.pb.gw.go: api/legacy/bbsim.proto setup_tools
	@echo $@
	@${PROTOC} -I. \
	  -I${GOOGLEAPI}/third_party/googleapis/ \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
      --go_out=plugins=grpc:./ \
	  --grpc-gateway_out=logtostderr=true,paths=source_relative,allow_delete_body=true:./ \
      $<

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docs/swagger/bbsim/bbsim.swagger.json: api/bbsim/bbsim.yaml setup_tools
	@echo $@
	@${PROTOC} -I ./api \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
	  --swagger_out=logtostderr=true,allow_delete_body=true,disable_default_errors=true,grpc_api_configuration=$<:docs/swagger/ \
	  api/bbsim/bbsim.proto

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
docs/swagger/leagacy/bbsim.swagger.json: api/legacy/bbsim.proto setup_tools
	@echo $@
	@${PROTOC} -I ./api \
	  -I${GOOGLEAPI}/ \
	  -I${VOLTHA_PROTOS}/protos/ \
	  --swagger_out=logtostderr=true,allow_delete_body=true,disable_default_errors=true:docs/swagger/ \
	  $<

## -----------------------------------------------------------------------
## Intent:
## -----------------------------------------------------------------------
clean ::
	$(RM) -r "docs/$(venv-name)"

## -----------------------------------------------------------------------
## Intent: Helper target used to exercise release script changes
## -----------------------------------------------------------------------
include $(MAKEDIR)/release/onf-publish.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
