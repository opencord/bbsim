# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2024 Open Networking Foundation (ONF) and the ONF Contributors
# -----------------------------------------------------------------------
# https://gerrit.opencord.org/plugins/gitiles/onf-make
# ONF.makefile.version = 1.1
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

help ::
	@echo
	@echo "[LINT]"

include $(legacy-mk)/lint/doc8/include.mk
include $(legacy-mk)/lint/docker/include.mk
include $(legacy-mk)/lint/groovy/include.mk
include $(legacy-mk)/lint/jjb.mk
include $(legacy-mk)/lint/json.mk
include $(legacy-mk)/lint/license/include.mk
include $(legacy-mk)/lint/makefile.mk
include $(legacy-mk)/lint/python/include.mk
include $(legacy-mk)/lint/shell.mk
include $(legacy-mk)/lint/yaml.mk

include $(legacy-mk)/lint/help.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
