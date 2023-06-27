# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2023 Open Networking Foundation (ONF) and the ONF Contributors
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

##--------------------------------##
##---]  Disable lint targets  [---##
##--------------------------------##
# NO-LINT-DOCKER        := true
NO-LINT-DOC8          := true

NO-LINT-GROOVY        := true#               # tests/bbr.groovy
NO-LINT-JJB           := not-relevant
# NO-LINT-JSON          := true#               # Note[1]
NO-LINT-MAKEFILE      := true
NO-LINT-ROBOT         := true
NO-LINT-SHELL         := true#               # Note[1]
NO-LINT-YAML          := true#               # Note[1]

NO-LINT-LICENSE       := true
# NO-LINT-REUSE         := true                # License check

NO-LINT-PYTHON        := true#               # Note[1]
# NO-LINT-PYTHON-FLAKE8 := true
# NO-LINT-PYTHON-PYLINT := true

# Note[1] - Plenty of source to cleanup
# Note[2] - No sources available

##-----------------------------------------##
##---]  Hack(s) to enable refactoring  [---##
##-----------------------------------------##
# USE_LINT_LICENSE   := voltha-system-tests

##---------------------------------##
##---] Conditional make logic  [---##
##---------------------------------##
# USE_DOCKER_MK    := true

# [EOF]
