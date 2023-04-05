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

YAML_FILES ?= $(error YAML_FILES= is required)

.PHONY: lint-yaml

lint : lint-yaml

lint-yaml: vst_venv
	source ./$</bin/activate \
	    ; set -u \
	    ; yamllint -s $(YAML_FILES)

help::
	@echo "  lint-yaml            Syntax check yaml source using yamllint"

# [EOF]
