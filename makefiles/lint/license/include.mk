# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022 Open Networking Foundation
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

# JSON_FILES ?= $(error JSON_FILES= is required)

.PHONY: lint-license

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint : lint-license

lint-license-gargs += --recursive

# ignore: png, xlsx
# will utf8 be excluded(?)
lint-license-gargs += --binary-files=without-match
lint-license-gargs += --files-without-match

# [TODO] license checking accepts either Copy or Apache.
# [TODO] At least Copyright should be required (both?)
lint-license-gargs += --extended-regexp
lint-license-gargs += -e 'Copyright[[:space:]]+[[:digit:]]{4}'
lint-license-gargs += -e 'Apache License'

# [TODO] --strict, --strict-dates

# Editor temp files
lint-license-gargs += --exclude='*.~'
lint-license-gargs += --exclude='*.swp'

lint-license-gargs += --exclude-dir='.git'

# TODO: Normalize into .venv for consistent filtering across projects.
lint-license-gargs += --exclude-dir='vst_venv'
lint-license-gargs += --exclude-dir='flog'

lint-license-gargs += --exclude='*.json'
lint-license-gargs += --exclude='*.md'
lint-license-gargs += --exclude='*.out'
lint-license-gargs += --exclude='*.pyc'
lint-license-gargs += --exclude='*.xml'

# [FILE(s)]
lint-license-gargs += --exclude='VERSION'

# [GIT]
# lint-license-gargs += --exclude='.gitignore'
# lint-license-gargs += --exclude='.gitreview'
lint-license-gargs += --exclude='\.*'

# [PYTHON]
lint-license-gargs += --exclude='requirements.txt'

# [WIP]
lint-license-gargs += --exclude='patch'

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
lint-license-new:
	grep $(lint-license-gargs) $(dot)

## -----------------------------------------------------------------------
## Jenkins job checking logic.
## -----------------------------------------------------------------------
lint-license:
	$(MAKEDIR)/lint/license/license-check.sh

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
help ::
	@echo "  lint-license         Verify sources contain a license block."
	@echo "  lint-license-new     Grep driven replacement logic."

# [EOF]
