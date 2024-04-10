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

## -----------------------------------------------------------------------
## Define vars based on relative import (normalize symlinks)
## Usage: include makefiles/onf/include.mk
## -----------------------------------------------------------------------
onf-mk-abs    ?= $(abspath $(lastword $(MAKEFILE_LIST)))
onf-mk-top    := $(subst /include.mk,$(null),$(onf-mk-abs))

include $(legacy-mk)/consts.mk
include $(legacy-mk)/help/include.mk       # render target help
include $(legacy-mk)/utils/include.mk      # dependency-less helper macros
include $(legacy-mk)/etc/include.mk        # banner macros

include $(legacy-mk)/virtualenv.mk#        # lint-{jjb,python} depends on venv

include $(legacy-mk)/lint/include.mk

# include $(legacy-mk)/git-submodules.mk
# include $(legacy-mk)/gerrit/include.mk
include $(legacy-mk)/golang/include.mk

include $(legacy-mk)/todo.mk
include $(legacy-mk)/help/variables.mk

##---------------------##
##---]  ON_DEMAND  [---##
##---------------------##
$(if $(USE_DOCKER_MK),$(eval $(legacy-mk)/docker/include.mk))

##-------------------##
##---]  TARGETS  [---##
##-------------------##
include $(legacy-mk)/targets/clean.mk
# include $(legacy-mk)/targets/check.mk
include $(legacy-mk)/targets/sterile.mk
# include $(legacy-mk)/targets/test.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
