/*
 * Copyright 2020-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"github.com/opencord/bbsim/api/bbsim"
	"strings"
)

var OLTAlarms = make(map[string]bbsim.AlarmType_Types)

func init() {
	for id, name := range bbsim.AlarmType_Types_name {
		if id := bbsim.AlarmType_Types(id); strings.HasPrefix(name, "ONU_") && id != bbsim.AlarmType_ONU_ALARM {
			OLTAlarms[name] = id
		}
	}
}

var ONUAlarms = map[string]bbsim.AlarmType_Types{
	"OLT_LOS": bbsim.AlarmType_LOS,
}
