#!/bin/bash
# -----------------------------------------------------------------------
# Copyright 2022 Open Networking Foundation
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

declare -i status=$#

while [ $# -gt 0 ]; do
    arg="$1"; shift
    echo "ERROR: Detected missing license header: ${arg}"
done

if [ $status -ne 0 ]; then
    exit 1
fi

/bin/true # set $?

# [EOF]
