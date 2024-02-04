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

##-------------------##
##---]  GLOBALS  [---##
##-------------------##

HADOLINT = $(docker-run-app) $(is-stdin) $(vee-citools)-hadolint hadolint

ifdef LOCAL_LINT
  lint-hadolint-dep = lint-hadolint-local
else
  lint-hadolint-dep = lint-hadolint
endif

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: $(lint-hadolint-dep)

lint : $(lint-hadolint-dep)

lint-dockerfile : $(lint-hadolint-dep)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint-hadolint:

	@echo
	@echo '** -------------------------------------------------------------'
	@echo "** $(MAKE): processing target [$@]"
	@echo '** -------------------------------------------------------------'
	$(HIDE)${HADOLINT} $$(find ./build -name "Dockerfile*")
	@echo "Dockerfile lint check OK"

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
hadolint-cmd := ./hadolint-Linux-x86_64

lint-hadolint-local: hadolint-get
	$(hadolint-cmd) $$(find ./build -name "Dockerfile*")

## -----------------------------------------------------------------------
## Intent: Retrieve the hadolint tool
## https://github.com/hadolint/hadolint/releases/tag/v2.12.0
## -----------------------------------------------------------------------
hadolint-get:
	true
#	$(MAKECMDGOALS)/lint/docker/get.sh
#	$(GIT) clone https://github.com/hadolint/hadolint.git
#	wget https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-x86_64

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@echo '  lint-dockerfile      Perform all dockerfile lint checks'
	@echo '  lint-hadolint        Dockerfile lint check'

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
todo ::
	@echo '  o Update lint-dockerfile to run all dockerfile lint checks'

# [SEE ALSO]
# https://github.com/hadolint/hadolint

# [EOF]
