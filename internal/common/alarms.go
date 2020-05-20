// Copyright 2020-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
