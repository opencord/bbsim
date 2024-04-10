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
TOP-docs ?=$(strip \
  $(dir \
    $(abspath $(lastword $(MAKEFILE_LIST)))\
   )\
)
TOP-docs := $(subst /docs/,$(null),$(TOP-docs))

##--------------------##
##---]  INCLUDES  [---##
##--------------------##
include $(TOP-docs)/config.mk
include $(TOP-docs)/makefiles/include.mk

# Minimal makefile for Sphinx documentation

# You can set these variables from the command line.
SPHINXOPTS    =
SPHINXBUILD   = sphinx-build
SOURCEDIR     = source
BUILDDIR      = build

.PHONY: help Makefile lint

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
clean ::
	$(RM) -r "$(venv-name)"
#	$(RM) -r build "$(venv-name)"

## -----------------------------------------------------------------------
## Intent: Display target help
## -----------------------------------------------------------------------
help :: $(venv-activate-script)
	$(activate) \
	  && $(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

# -----------------------------------------------------------------------
# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
# -----------------------------------------------------------------------
docs-targets += html

$(docs-targets) : $(venv-activate-script) Makefile
	$(call banner-enter,Target $@)
	$(activate) \
	  && $(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)
	$(call banner-leave,Target $@)

# [EOF]
