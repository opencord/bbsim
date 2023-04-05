# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2023 Open Networking Foundation
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

GOLANG_FILES ?= $(error PYTHON_FILES= is required)

.PHONY: lint-golang-sca

lint : lint-golang-sca

## -----------------------------------------------------------------------
## Intent: Run goformat on files on sandbox files.
##   1) find . -name '*.go' -print0
##      - gather all *.go sources (-name '*.go')
##      - pass as a list of null terminated items (-print0)
##   2) xargs --null --max-args=[n] --no-run-if-empty gofmt -d
##      - Iterate over the list (xargs --null)
##      - process one item per line (--max-args=1)
##      - display filename-to-check (--verbose)
##      - display content when diffs are detected:
##           gofmt -d
##           gofmt -d -s
## -----------------------------------------------------------------------
lint-golang-sca-xargs := $(null)
lint-golang-sca-xargs += --null#+           # Source paths are null terminated
lint-golang-sca-xargs += --max-args=1#+     # Check one file at a time
lint-golang-sca-xargs += --no-run-if-empty
lint-golang-sca-xargs += --verbose#+        # Display source path to check

## [INPLACE-EDITS] make lint-golang-sca FIX=1
ifdef FIX
  lint-golang-sca-args += -w
endif

lint-golang-sca:
	find . -name '*.go' -print0 \
	    | xargs $(lint-golang-sca-xargs) gofmt -d -s

help::
	@echo "  lint-golang-sca            Syntax check golang sources"
	@echo "    MODIFIER: FIX=1          Correct problems (gofmt -d -s -w)"

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
