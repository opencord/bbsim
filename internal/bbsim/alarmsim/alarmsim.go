/*
 * Copyright 2018-present Open Networking Foundation

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

package alarmsim

import (
	"context"
	"fmt"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"strconv"
)

// Find a key in the optional AlarmParameters, convert it to an integer,
// return 'def' if no key exists or it cannot be converted.
func extractInt(params []*bbsim.AlarmParameter, name string, def int) int {
	for _, kv := range params {
		if kv.Key == name {
			i, err := strconv.Atoi(kv.Value)
			if err == nil {
				return i
			}
		}
	}
	return def
}

func BuildAlarmIndication(req *bbsim.ONUAlarmRequest, o *devices.OltDevice) (*openolt.AlarmIndication, error) {
	var alarm *openolt.AlarmIndication

	onu, err := o.FindOnuBySn(req.SerialNumber)
	if err != nil {
		return nil, err
	}

	switch req.AlarmName {
	case "DyingGasp":
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_DyingGaspInd{&openolt.DyingGaspIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case "StartupFailure":
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuStartupFailInd{&openolt.OnuStartupFailureIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case "SignalDegrade":
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuSignalDegradeInd{&openolt.OnuSignalDegradeIndication{
				Status:              req.Status,
				OnuId:               onu.ID,
				IntfId:              onu.PonPortID,
				InverseBitErrorRate: uint32(extractInt(req.Parameters, "InverseBitErrorRate", 0)),
			}},
		}
	default:
		return nil, fmt.Errorf("Unknown alarm name %s", req.AlarmName)
	}

	return alarm, nil
}

func SimulateAlarm(ctx context.Context, req *bbsim.ONUAlarmRequest, o *devices.OltDevice) error {
	alarmIndication, err := BuildAlarmIndication(req, o)
	if err != nil {
		return err
	}

	/*msg := &openolt.Indication_AlarmInd{AlarmInd: alarm}
	ind := &openolt.Indication{Data: msg}
	*/

	err = o.SendAlarmIndication(ctx, alarmIndication)
	if err != nil {
		return err
	}

	return nil
}
