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

help::
	@echo
	@echo "[LINT]"
	@echo "  lint                       Static code analysis/syntax checking"
	@echo "    LOCAL_LINT=1             Enable local linting w/o docker & jenkins overhead"

include $(MAKEDIR)/lint/docker/include.mk

ifdef LOCAL_LINT
  # include $(MAKEDIR)/lint/json.mk              # venv dependency
  # include $(MAKEDIR)/lint/golang/sca.mk
  # include $(MAKEDIR)/lint/license/include.mk   # exclusions needed
  # include $(MAKEDIR)/lint/python.mk
  # include $(MAKEDIR)/lint/robot.mk             # venv dependency
  # include $(MAKEDIR)/lint/shell.mk             # cleanup needed
  # include $(MAKEDIR)/lint/yaml.mk              # venv needed -- alt: yamllint
endif

# [EOF]
