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

ROBOT_FILES ?= $(error ROBOT_FILES= is required)

LINT_ARGS ?= --verbose --configure LineTooLong:130 -e LineTooLong \
             --configure TooManyTestSteps:65 -e TooManyTestSteps \
             --configure TooManyTestCases:50 -e TooManyTestCases \
             --configure TooFewTestSteps:1 \
             --configure TooFewKeywordSteps:1 \
             --configure FileTooLong:2000 -e FileTooLong \
             -e TrailingWhitespace


.PHONY: lint-robot

lint : lint-robot

lint-robot: $(venv-activate-script)
	$(activate) && rflint $(LINT_ARGS) $(ROBOT_FILES)

help::
	@echo "  lint-robot           Syntax check robot sources using rflint"

# [EOF]
