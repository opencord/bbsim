# -*- makefile -*-
# -----------------------------------------------------------------------
# Copyright 2017-2023 Open Networking Foundation (ONF) and the ONF Contributors
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

onf-publish-args := $(null)
onf-publish-args += --draft
onf-publish-args += --gen-version
onf-publish-args += --repo-org opencord

# args+=('--repo' 'github.com/opencord/bbsim')
args+=('--title' 'bbsim - v9876.543.21')

onf-publish-args += --repo-name bbsim
onf-publish-args += --git-hostname github.com
onf-publish-args += --pac $(HOME)/.ssh/github.com/pacs/onf-voltha
# onf-publish-args += --notes-file "$(PWD)/release-notes"

onf-publish-files += release/bbsimctl-windows-amd64
onf-publish-files += release/bbr-linux-amd64
onf-publish-files += release/bbr-windows-amd64
onf-publish-files += release/bbr-darwin-amd64
onf-publish-files += release/bbr-linux-arm64
onf-publish-files += release/bbsim-linux-arm64
onf-publish-files += release/bbsimctl-darwin-amd64
onf-publish-files += release/checksum.SHA256
onf-publish-files += release/bbsimctl-linux-amd64
onf-publish-files += release/bbsim-darwin-amd64
onf-publish-files += release/bbsim-windows-amd64
onf-publish-files += release/bbsim-linux-amd64
onf-publish-files += release/bbsimctl-linux-arm64

onf-publish:
	../ci-management/jjb/shell/github-release.sh $(onf-publish-args) 2>&1 | tee $@.log

# [EOF]
