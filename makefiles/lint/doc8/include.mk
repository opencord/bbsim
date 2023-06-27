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

##-------------------##
##---]  GLOBALS  [---##
##-------------------##
.PHONY: lint-doc8 lint-doc8-all lint-doc8-modified

have-rst-files := $(if $(strip $(RST_SOURCE)),true)
RST_SOURCE     ?= $(error RST_SOURCE= is required)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
ifndef NO-LINT-DOC8
  lint-doc8-mode := $(if $(have-doc8-files),modified,all)
  lint : lint-doc8-$(lint-doc8-mode)
endif# NO-LINT-DOC8

# Consistent targets across lint makefiles
lint-doc8-all      : lint-doc8
lint-doc8-modified : lint-doc8

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
include $(MAKEDIR)/lint/doc8/excl.mk

ifdef lint-doc8-excl
  lint-doc8-excl-args += $(addprefix --ignore-path$(space),$(lint-doc8-excl))
endif
lint-doc8-excl-args += $(addprefix --ignore-path$(space),$(lint-doc8-excl-raw))

lint-doc8-args += --max-line-length 120

lint-doc8: $(venv-activate-script)
	@echo
	@echo '** -----------------------------------------------------------------------'
	@echo '** doc8 *.rst syntax checking'
	@echo '** -----------------------------------------------------------------------'
	$(activate) && doc8 --version
	@echo
	$(activate) && doc8 $(lint-doc8-excl-args) $(lint-doc8-args) .

## -----------------------------------------------------------------------
## Intent: Display command usage
## -----------------------------------------------------------------------
help::
	@echo '  lint-doc8          Syntax check python using the doc8 command'
  ifdef VERBOSE
	@echo '  lint-doc8-all       doc8 checking: exhaustive'
	@echo '  lint-doc8-modified  doc8 checking: only modified'
  endif

# [EOF]
