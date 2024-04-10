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
#
# SPDX-FileCopyrightText: 2024 Open Networking Foundation (ONF) and the ONF Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(legacy-mk)/release/consts.mk
include $(legacy-mk)/release/macros.mk

## -----------------------------------------------------------------------
## Intent: Dispaly a help section for release targets
## -----------------------------------------------------------------------
help::
	@echo
	@echo '[RELEASE]'
	@echo '  release             Creates release ready binaries for BBSimctl and BBR artifacts'

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(legacy-mk)/release/bbr.mk
include $(legacy-mk)/release/bbsim.mk
include $(legacy-mk)/release/bbsimctl.mk

## -----------------------------------------------------------------------
## Intent: Cross-compile binaries for release
## -----------------------------------------------------------------------
release-target-deps += release-bbr
release-target-deps += release-bbsim
# release-deps += release-bbsimctl

## -----------------------------------------------------------------------
## Intent: Create a release from targets release-{bbr,bbsim,bbsimctl}
## Debug: $(foreach tgt,$(release-target-deps),$(info ** release-target-deps=$(tgt)))
## -----------------------------------------------------------------------
.PHONY: release
release: $(release-target-deps)

## -----------------------------------------------------------------------
## Intent: Initialize storage then build release target groups (bbr, bbsim)
## -----------------------------------------------------------------------
.PHONY: $(release-targetdeps)
$(release-target-deps) : release-init-mkdir

## -----------------------------------------------------------------------
## Intent: Prepare for a clean release build
## -----------------------------------------------------------------------
release-init:
	@echo
	@echo "** -----------------------------------------------------------------------"
	@echo "** $(MAKE): processing target [$@]"
	@echo "** -----------------------------------------------------------------------"

	@echo -e "\n** Create filesystem target for docker volume: $(RELEASE_DIR)"
	$(RM) -r "./$(RELEASE_DIR)"
	mkdir -vp "$(RELEASE_DIR)"

## -----------------------------------------------------------------------
## Intent: Create storage for release binaries and docker volume mount
## -----------------------------------------------------------------------
.PHONY: release-init
release-init-mkdir:
	mkdir -vp "$(RELEASE_DIR)"

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
