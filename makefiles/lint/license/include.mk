# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2024 Open Networking Foundation (ONF) and the ONF Contributors
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

ifndef NO-LINT-LICENSE

ifndef mk-include--onf-lint-license#       # one-time loader

$(if $(DEBUG),$(warning ENTER))

$(if $(USE_LINT_LICENSE)\
  ,$(eval include $(legacy-mk)/lint/license/voltha-system-tests/include.mk)\
  ,$(eval include $(legacy-mk)/lint/license/common.mk)\
)

  mk-include--onf-lint-license := true

$(if $(DEBUG),$(warning LEAVE))

endif # mk-include--onf-lint-license

endif # NO-LINT-LICENSE

# [EOF]
