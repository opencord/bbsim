# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2023 Open Networking Foundation (ONF) and the ONF Contributors
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

ifndef mk-include--onf-lint-python#       # one-time loader

$(if $(DEBUG),$(warning ENTER))

## -----------------------------------------------------------------------
## Intent: Display early so lint targets are grouped
## -----------------------------------------------------------------------
help ::
#	@echo
#	@echo '[PYTHON]'
	@echo '  lint-python         Syntax check python sources (*.py)'
#	@echo '  help-lint-python-flake8'
#	@echo '  help-lint-python-pylint'

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
  ifndef NO-LINT-PYTHON
    include $(ONF_MAKEDIR)/lint/python/flake8.mk
    include $(ONF_MAKEDIR)/lint/python/pylint.mk
  endif

  mk-include--onf-lint-python := true

$(if $(DEBUG),$(warning LEAVE))

endif # mk-include--onf-lint-license

# [EOF]
