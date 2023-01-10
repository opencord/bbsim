# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2023 Open Networking Foundation (ONF) and the ONF Contributors (ONF) and the ONF Contributors
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

## -----------------------------------------------------------------------
## -----------------------------------------------------------------------
.PHONY: help
help :: # @HELP Print the command options
	@echo
	@echo -e "\033[0;31m    BroadBand Simulator (BBSim) \033[0m"
	@echo
	@echo Emulates the control plane of an openolt compatible device
	@echo Useful for development and scale testing
	@echo
	@grep -E '^.*: .* *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": .* *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '

.PHONY: HELP
HELP: # @HELP Display extended makefile target help
	@echo "Usage: $(MAKE) [options] [target] ..."
	@echo
	@echo "[DOCS]          Generate documentation"
	@echo "  docs          (alias) swagger"
	@echo "  swagger       Generate swagger documentation for BBSIM API"
	@echo
	@echo "[LINT] - Invoke source linters"
	@echo "  docs-lint     Lint generated docs"
	@echo "  lint          (alias) lint-dockerfile, lint-mod"
	@echo "  lint-dockerfile"
	@echo "  lint-mod      Golang module verify"
	@echo "  sca           Static code analysis"
	@echo
	@echo "[SETUP]"
	@echo "  setup_tools   Install module dependencies"
	@echo
	@echo "[BUILD]"
	@echo
	@echo "[RELEASE]"
	@echo
	@echo "[CLEAN]"
	@echo "  clean         Remove generated targets"
	@echo

## -----------------------------------------------------------------------
## ----------------------------------------------------------------------
.PHONY: help-all
help-all : help HELP

# [EOF]
