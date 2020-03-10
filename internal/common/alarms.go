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
