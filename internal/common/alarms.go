/*
 * Copyright 2020-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
)

var ONUAlarms = make(map[string]bbsim.AlarmType_Types)

func init() {
	for id, name := range bbsim.AlarmType_Types_name {
		if id := bbsim.AlarmType_Types(id); id != bbsim.AlarmType_LOS {
			ONUAlarms[name] = id
		}
	}
}

const OltNniLos = "OLT_NNI_LOS"
const OltPonLos = "OLT_PON_LOS"

var OLTAlarms = map[string]bbsim.AlarmType_Types{
	OltNniLos: bbsim.AlarmType_LOS,
	OltPonLos: bbsim.AlarmType_LOS,
}
