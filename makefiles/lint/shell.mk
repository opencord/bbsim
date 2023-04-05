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

JSON_FILES ?= $(error JSON_FILES= is required)

.PHONY: lint-shell

lint : lint-shell

lint-shell:
	shellcheck --version
	find . \( -name 'staging' -o -name 'vst_venv' \) -prune \
	    -o -name '*.sh' ! -name 'activate.sh' -print0 \
	| xargs -0 -n1 shellcheck

help::
	@echo "  lint-shell           Syntax check bash,bourne,etc sources"

# [EOF]
