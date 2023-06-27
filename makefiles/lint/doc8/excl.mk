# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2022 Open Networking Foundation
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
## -----------------------------------------------------------------------

## excl := $(wildcar */*/.git)
lint-doc8-excl-raw += '$(venv-name)'
lint-doc8-excl-raw += '*/$(venv-name)'
$(if $(BUILDDIR),\
  $(excl lint-doc8-excl-raw += '$(BUILDDIR)'))

# YUCK! -- overhead
#   o Submodule(s) use individual/variant virtualenv install paths.
#   o Exclude special snowflakes to enable library makefile use.
#   o All can use virtualenv.mk for consistent names and cleanup
#   o [TODO] Ignore submodules, individual repos should check their sources.

lint-doc8-excl-raw += '*/venv_cord'

lint-doc8-excl-raw += './cord-tester'
lint-doc8-excl-raw += './repos/cord-tester'

lint-doc8-excl-raw += './bbsim/internal/bbsim/responders/sadis/dp.txt'
lint-doc8-excl-raw += './repos/bbsim/internal/bbsim/responders/sadis/dp.txt'
lint-doc8-excl-raw += './repos/voltha-helm-charts/voltha-infra/templates/NOTES.txt'
lint-doc8-excl-raw += './voltha-helm-charts/voltha-infra/templates/NOTES.txt'

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
