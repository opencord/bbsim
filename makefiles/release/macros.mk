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

$(if $(DEBUG),$(warning ENTER))

## -----------------------------------------------------------------------
## Intent: Given a hyphen separated list of values return item [n]
## -----------------------------------------------------------------------
## Usage:
##   value    := foo-bar-tans-fans
##   tans-val = $(call get-hyphen-field,value,3)
## -----------------------------------------------------------------------
get-hyphen-field = $(word $(1),$(subst -,$(space),$(notdir $(2))))

## -----------------------------------------------------------------------
## Intent: Construct a parameterized $(GO_SH) command
## Given:
##   argv[1] Response file containing environment vars passed to docker.
## Usage:
##   cmd = $(my-go-sh,./env-vars)
## Debug: view arguments passed
##   make DEBUG=1
## -----------------------------------------------------------------------
my-go-sh=$(strip                   \
  $(docker-run-app)                \
	$(is-stdin)                \
	$(if $(1),--env-file $(1)) \
	-v gocache:/.cache         \
	$(vee-golang)              \
	$(vee-citools)-golang      \
	sh -c                      \
)

## -----------------------------------------------------------------------
## Intent: Derive a list of release platform binaries by name
## -----------------------------------------------------------------------
## Given:
##   $1 (scalar:name)      $(RELEASE_BBSIM_NAME)
##   $2 (indirect:arches)  $(RELEASE_OS_ARCH)
## -----------------------------------------------------------------------
## Usage: deps = $(call release-gen-deps-name,tool-name,tool-arches)
## -----------------------------------------------------------------------
release-gen-deps-name=$(strip \
\
  $(if $($(1)),$(null),$(error name= is required))\
  $(if $($(2)),$(null),$(error arches= is required))\
\
  $(foreach name,$($(1)),\
     $(if $(DEBUG),$(info name=$(name)))\
  $(foreach arches,$(2),\
     $(if $(DEBUG),$(info arches=$(arches)))\
     $(foreach arch,$($(arches)),\
        $(if $(DEBUG),$(info arch=$(arch): $(name)-$(arch)))\
        $(name)-$(arch)\
     )\
  ))\
)

## -----------------------------------------------------------------------
## Intent: Derive a list of release binary dependencies
## -----------------------------------------------------------------------
## Returns:
##    release/bbsim-darwin-amd64
##    release/bbsim-linux-amd64
##    release/bbsim-windows-amd64
## -----------------------------------------------------------------------
## Usage:
##    tool-name   = bbsim
##    tool-arches = darwin linux windows
##    release-dir = release
##    deps = $(call release-gen-deps,tool-name,tool-arches,release-dir)
##    $(foreach val,$(deps),$(info ** deps=$(val)))
## -----------------------------------------------------------------------
release-gen-deps=$(strip \
\
  $(if $($(1)),$(null),$(error name= is required))\
  $(if $($(2)),$(null),$(error arches= is required))\
  $(if $($(3)),$(null),$(error release-dir= is required))\
\
  $(if $(DEBUG),$(info release-dir=$(3)))\
  $(foreach release-dir,$($(3)),\
    $(addprefix $(release-dir)/,$(call release-gen-deps-name,$(1),$(2)))\
  )\
)

$(if $(DEBUG),$(warning LEAVE))

# [EOF]
