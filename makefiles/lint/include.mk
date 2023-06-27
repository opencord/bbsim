# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2022-2023 Open Networking Foundation (ONF) and the ONF Contributors
# -----------------------------------------------------------------------
# https://gerrit.opencord.org/plugins/gitiles/onf-make
# ONF.makefile.version = 1.1
# -----------------------------------------------------------------------

$(if $(DEBUG),$(warning ENTER))

help ::
	@echo
	@echo "[LINT]"

include $(ONF_MAKEDIR)/lint/doc8/include.mk
include $(ONF_MAKEDIR)/lint/docker/include.mk
include $(ONF_MAKEDIR)/lint/groovy/include.mk
include $(ONF_MAKEDIR)/lint/jjb.mk
include $(ONF_MAKEDIR)/lint/json.mk
include $(ONF_MAKEDIR)/lint/license/include.mk
include $(ONF_MAKEDIR)/lint/makefile.mk
include $(ONF_MAKEDIR)/lint/python/include.mk
include $(ONF_MAKEDIR)/lint/shell.mk
include $(ONF_MAKEDIR)/lint/yaml.mk

include $(ONF_MAKEDIR)/lint/help.mk

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
