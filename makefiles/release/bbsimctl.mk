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
RELEASE_BBSIMCTL_NAME ?= bbsimctl

bbsimctl-config := github.com/opencord/bbsim/internal/bbsimctl/config

# release-target-deps += release-bbsimctl

# -----------------------------
# release/bbsimctl-darwin-amd64
# release/bbsimctl-linux-amd64
# release/bbsimctl-windows-amd64
# -----------------------------
release-bbsimctl-deps :=\
  $(call release-gen-deps,RELEASE_BBSIMCTL_NAME,RELEASE_OS_ARCH,RELEASE_DIR)

## -----------------------------------------------------------------------
## Intent: Cross-compile bbsimctl binaries as dependency targets
##   o target: release/bbsimctl-linux-amd64
##   o create release/ for VOLUME mounting by docker container
##   o create a response file for passing docker env vars
##   o cross-compile: GOOS= GOARCH= go build
## -----------------------------------------------------------------------
release-bbsimctl: $(release-bbsimctl-deps)

.PHONY: $(release-bbsimctl-deps)
$(release-bbsimctl-deps):

	@echo
	@echo "** -----------------------------------------------------------------------"
	@echo "** $(MAKE): processing target [$@]"
	@echo "** -----------------------------------------------------------------------"

        # Docker container is responsible for compiling
        # Release target will publish from localhost:release/
        # Binaries are built into mounted docker volume /app/release=> localhost:release/
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
	 -ldflags "-w -X $(bbsimctl-config).BuildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X $(bbsimctl-config).CommitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X $(bbsimctl-config).GitStatus=${GIT_STATUS} \
	    -X $(bbsimctl-config).Version=${VERSION}" \
	  -o "$@" ./cmd/bbsimctl \
    $(quote-single)

        # -----------------------------------------------------------------------
        # Cleanup and display results
        # -----------------------------------------------------------------------
	@$(RM) $(notdir $@).env
	$(HIDE)umask 000 && chmod 755 "$(RELEASE_DIR)"
	$(HIDE)file "$@"

## -----------------------------------------------------------------------
## Intent: Build bbsimctl on localhost
## -----------------------------------------------------------------------
build-bbsimctl:
	@go build -mod vendor \
	  -ldflags "-w -X github.com/opencord/bbsim/internal/bbsimctl/config.BuildTime=$(shell date +%Y/%m/%d-%H:%M:%S) \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.CommitHash=$(shell git log --pretty=format:%H -n 1) \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.GitStatus=${GIT_STATUS} \
	    -X github.com/opencord/bbsim/internal/bbsimctl/config.Version=${VERSION}" \
	  ./cmd/bbsimctl

## -----------------------------------------------------------------------
## Intent: Remove generated targets
## -----------------------------------------------------------------------
clean ::
	$(RM) $(release-bbsimctl-deps)
	$(RM) *.env
	$(RM) bbsimclt

## -----------------------------------------------------------------------
## Intent: Display target help#
# -----------------------------------------------------------------------
help ::
	@echo
	@echo '[bbsimctl]'
	@echo '  build-bbsimctl      Compile bbsimctl on localhost'
	@echo '  release-bbsimctl    Cross-compile bbsimctl binaries into a docker mounted filesystem'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
