#!/bin/bash
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
#
# SPDX-FileCopyrightText: 2023 Open Networking Foundation (ONF) and the ONF Contributors
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------

## -----------------------------------------------------------------------
## Intent: Install a bbsim binary for local development use.
## -----------------------------------------------------------------------
## Note: A python or golang script may be a simpler answer.
##       Interpreter modules provide answers for uname -{a,m,o}
##       with dictionary translation into needed values.
## -----------------------------------------------------------------------

# import platform
# platform.processor()
# platform.system        # Windows

# lshw:    width: 64 bits

# >>> import platform
# >>> platform.machine()
# 'x86'


# $ uname -m
# armv7l

## which arch
# https://github.com/hadolint/hadolint/releases/tag/v2.12.0
case "$(uname -a)" in
    *x86_64*)
esac

os=''
case "$(uname -o)" in
    *Linux*) os='Linux'
esac


# hadolint-Darwin-x86_64
# hadolint-Darwin-x86_64.sha256
# hadolint-Linux-arm64
# hadolint-Linux-arm64.sha256
# hadolint-Linux-x86_64
# hadolint-Linux-x86_64.sha256
# hadolint-Windows-x86_64.exe
# hadolint-Windows-x86_64.exe.sha256
# Source code (zip)
# Source code (tar.gz)
