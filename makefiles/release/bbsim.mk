# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2024 Open Networking Foundation (ONF) and the ONF Contributors
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

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
RELEASE_BBSIM_NAME    ?= bbsim

# release-target-deps += release-bbsim

# -----------------------------
# release/bbsim-darwin-amd64
# release/bbsim-linux-amd64
# release/bbsim-windows-amd64
# -----------------------------
release-bbsim-deps :=\
  $(call release-gen-deps,RELEASE_BBSIM_NAME,RELEASE_OS_ARCH,RELEASE_DIR)

## -----------------------------------------------------------------------
## Intent: Cross-compile bbsim binaries as dependency targets
##   o target: release/bbsim-linux-amd64
##   o create release/ for VOLUME mounting by docker container
##   o create a response file for passing docker env vars
##   o cross-compile: GOOS= GOARCH= go build
## -----------------------------------------------------------------------
release-bbsim: $(release-bbsim-deps)

.PHONY: $(release-bbsim-deps)
$(release-bbsim-deps):

	@echo
	@echo "** -----------------------------------------------------------------------"
	@echo "** $(MAKE): processing target [$@]"
	@echo "** -----------------------------------------------------------------------"

        # Docker container is responsible for compiling
        # Release target will publish from localhost:release/
        # Binaries are built into mounted docker volume /app/release => localhost:release/
	$(HIDE)mkdir -vp "$(RELEASE_DIR)"
	$(HIDE)umask 000 && chmod 777 "$(RELEASE_DIR)"

        # -----------------------------------------------------------------------
        # Create a response file for passing environment vars to docker
        # -----------------------------------------------------------------------
	$(HIDE)$(RM) $(notdir $@).env
	$(HIDE)echo -e '#!/bin/bash\n' \
	  'GOOS=$(call get-hyphen-field,2,$@)\n' \
	  'GOARCH=$(call get-hyphen-field,3,$@)\n' \
	  >> "$(notdir $@).env"

        # -----------------------------------------------------------------------
        # Compile a platform binary
        # -----------------------------------------------------------------------
	$(HIDE) \
	umask 022 \
\
	&& echo "** Building: $@" \
	&& set -x \
	&& $(call my-go-sh,$(notdir $@).env) \
    $(quote-single) \
      go build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  -o "$@" ./cmd/bbsim \
    $(quote-single)

        # -----------------------------------------------------------------------
        # Cleanup and display results
        # -----------------------------------------------------------------------
	@$(RM) $(notdir $@).env
	$(HIDE)umask 000 && chmod 755 "$(RELEASE_DIR)"
	$(HIDE)file "$@"


release-bbsim-x:
	@echo "$(RELEASE_BBSIM_NAME)-linux-amd64"
	${GO} build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  -o "$(RELEASE_DIR)/$(RELEASE_BBSIM_NAME)-linux-amd64" ./cmd/bbsim

## -----------------------------------------------------------------------
## Intent: Build bbsimctl on localhost
## -----------------------------------------------------------------------
build-bbsim:
	@go build -mod vendor \
	  -ldflags "-w -X main.buildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X main.commitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X main.gitStatus=${GIT_STATUS} \
	    -X main.version=${VERSION}" \
	  ./cmd/bbsim

## -----------------------------------------------------------------------
## Intent: Remove generated targets
## -----------------------------------------------------------------------
clean ::
	$(RM) $(release-bbsim-deps)
	$(RM) *.env
	$(RM) $(RELEASE_BBSIM_NAME)

## -----------------------------------------------------------------------
## Intent: Display target help#
# -----------------------------------------------------------------------
help ::
	@echo
	@echo '[bbsim]'
	@echo '  build-bbsim      Compile bbsim on localhost'
	@echo '  release-bbsim    Cross-compile bbsim binaries into a docker mounted filesystem'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
