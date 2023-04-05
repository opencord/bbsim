# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2019-2023 Open Networking Foundation
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

## -----------------------------------------------------------------------
## Intent:
##   o Generate volctl binaries in a docker container container
##   o Copy container:/apps/release to localhost:{pwd}/release
## -----------------------------------------------------------------------
## [TODO] Replace ${GO_SH} $(quote-single) ..   with $(call quoted,cmd-text)
## -----------------------------------------------------------------------
release-build :

	@echo 
	@echo "** -----------------------------------------------------------------------"
	@echo "** $(MAKE): processing target [$@]"
	@echo "** Sandbox: $(shell /bin/pwd)"
	@echo "** -----------------------------------------------------------------------"

        # [DEBUG] Yes this will take a while but where-4-art-thou-golang-in-docker-image-(?)
	@echo -e "\n** golang interpreter"
        # find '/usr/local' '/go' '/usr/bin' '/bin' -name 'go' ! -type d -print;
        # -${GO_SH} $(quote-single) find / -name 'go' ! -type d -print $(quote-single)

	-$(HIDE)${GO_SH} $(call quoted,which$(space)-a$(space)go)
	-$(HIDE)${GO_SH} $(call quoted,go$(space)version)

	@echo -e "\n** Create filesystem target for docker volume: $(RELEASE_DIR)"
	$(RM) -r "./$(RELEASE_DIR)"
	mkdir -vp "$(RELEASE_DIR)"

	$(MAKE) docker-debug

	@echo
	@echo '** -----------------------------------------------------------------------'
	@echo '** Filesystem: docker-container::/app  (wanted: release/)'
	@echo '** -----------------------------------------------------------------------'
	${GO_SH} $(quote-single)find /app \( -name ".git" -o -name "vendor" -o -name "makefiles" -o -name "internal" -o -name "pkg" \) -prune -o -print$(quote-single)
	@echo
	@echo '** /app/release permissions'
#	${GO_SH} $(quote-single)umask 022 && chmod 700 /app/release $(quote-single)
	${GO_SH} $(quote-single)/bin/ls -ld /app/release $(quote-single)

	@echo
	@echo '** -----------------------------------------------------------------------'
	@echo '** Docker builds bins into mounted filesystem:'
	@echo '**   container:/app/relase'
	@echo '**   localhost:{pwd}/release'
	@echo '** -----------------------------------------------------------------------'

#       NOTE: Use double quotes in echo strings else command breakage
	${GO_SH} $(quote-single) \
          echo ;\
	  echo "build: ENTER" ;\
	  set -e -o pipefail; \
	  set -x ; \
	  for x in ${RELEASE_OS_ARCH}; do \
	    echo ;\
	    echo "** RELEASE_OS_ARCH: Build arch is $$x"; \
	    OUT_PATH="$(RELEASE_DIR)/$(RELEASE_NAME)-$(subst -dev,_dev,$(VERSION))-$$x"; \
            echo ;\
            echo "** Building: $$OUT_PATH (ENTER)"; \
	    GOOS=$${x%-*} GOARCH=$${x#*-} go build -mod=vendor -v $(LDFLAGS) -o "$$OUT_PATH" cmd/voltctl/voltctl.go; \
            echo "** Building: $$OUT_PATH (LEAVE)"; \
	  done ;\
	  echo "build: LEAVE" \
$(quote-single)

	@echo
	@echo "** -----------------------------------------------------------------------"
	@echo '** Post-build, files to release'
	@echo "** -----------------------------------------------------------------------"
	-find "$(RELEASE_DIR)" ! -type d -print
	@echo

## -----------------------------------------------------------------------
## Intent: Why is go not found reported after
## -----------------------------------------------------------------------
docker-debug:

	@echo
	@echo "** -----------------------------------------------------------------------"
	@echo "** [TARGET] $@"
	@echo "** -----------------------------------------------------------------------"

	@echo
	docker images

	@echo
	docker ps --all

	@echo

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
# clean ::
#	$(RM) -r "./$(RELEASE_DIR)"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@echo '  release-build       Cross-compile binaries into a docker mounted filesystem'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
