# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2024 Open Networking Foundation Contributors
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
# SPDX-FileCopyrightText: 2017-2024 Open Networking Foundation Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------
# Intent: This makefile maintains file and directory exclusion lists
#         for doc8 linting.
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------

lint-doc8-excl-raw += '*/.git'
lint-doc8-excl-raw += '$(venv-name)'
lint-doc8-excl-raw += '*/$(venv-name)'
lint-doc8-excl-raw += './lf/onf-make'
lint-doc8-excl-raw += 'LICENSES/'

# Should we filter generated content as redundant ?
# Considering _build will be published and consumed.
$(if $(BUILDDIR),\
  $(excl lint-doc8-excl-raw += '$(BUILDDIR)'))

# -----------------------------------------------------------------------
# YUCK! -- overhead
#   o Submodule(s) use individual/variant virtualenv install paths.
#   o Exclude special snowflakes to enable library makefile use.
#   o All can use virtualenv.mk for consistent names and cleanup
#   o [TODO] Ignore submodules, individual repos should check their sources.
# -----------------------------------------------------------------------
lint-doc8-excl-raw += '*/venv_*'#   # '*/venv_cord'
lint-doc8-excl-raw += '*/*_venv'#   # '*/vst_venv'

## -----------------------------------------------------------------------
## repo:voltha-docs exclusions
## Migrate into one of:
##   makefiles/local/lint/doc8/doc8.ini
##   makefiles/local/lint/doc8/excl.mk
## -----------------------------------------------------------------------
lint-doc8-excl-raw += './cord-tester'
lint-doc8-excl-raw += './repos/cord-tester'

lint-doc8-excl-raw += './bbsim/internal/bbsim/responders/sadis/dp.txt'
lint-doc8-excl-raw += './repos/bbsim/internal/bbsim/responders/sadis/dp.txt'
lint-doc8-excl-raw += './repos/voltha-helm-charts/voltha-infra/templates/NOTES.txt'
lint-doc8-excl-raw += './voltha-helm-charts/voltha-infra/templates/NOTES.txt'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
