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
# SPDX-FileCopyrightText: 2024 Open Networking Foundation (ONF) and the ONF Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-update
mod-update: mod-tidy mod-vendor

## -----------------------------------------------------------------------
## Intent: Invoke the golang mod tidy command
## -----------------------------------------------------------------------
.PHONY: mod-tidy
mod-tidy :
	$(call banner-enter,Target $@)
	$(if $(LOCAL_FIX_PERMS),chmod o+w $(CURDIR))
	${GO} mod tidy
	$(if $(LOCAL_FIX_PERMS),chmod o-w $(CURDIR))
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: mod-vendor
mod-vendor : mod-tidy
mod-vendor :
	$(call banner-enter,Target $@)
	$(if $(LOCAL_FIX_PERMS),chmod o+w $(CURDIR))
	${GO} mod vendor
	$(if $(LOCAL_FIX_PERMS),chmod o-w $(CURDIR))
	$(call banner-leave,Target $@)

## -----------------------------------------------------------------------
## Intent: Display topic help
## Usage:
##   % make help
## -----------------------------------------------------------------------
help-summary ::
	@printf '  %-30s %s\n' 'mod-update' \
	  'Alias for make mod-tidy mod-update (GOLANG)'
  ifdef VERBOSE
	@$(MAKE) --no-print-directory mod-update-help
  endif

## -----------------------------------------------------------------------
## Intent: Display extended topic help
## Usage:
##   % make mod-update-help
##   % make help VERBOSE=1
## -----------------------------------------------------------------------
.PHONY: mod-update-help
mod-update-help ::
	@printf '  %-30s %s\n' 'mod-tidy'\
	  'Invoke go mod tidy'
	@printf '  %-30s %s\n' 'mod-vendor'\
	  'Invoke go mod vendor'
	@echo
	@echo '[MODIFIER]'
	@printf '  %-30s %s\n' 'LOCAL_FIX_PERMS=1' \
	  'Local dev hack to fix docker uid/gid volume problem'

# [EOF]
