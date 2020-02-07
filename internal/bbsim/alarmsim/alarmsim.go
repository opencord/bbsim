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

// AlarmNameMap string to enum map
var AlarmNameMap = map[string]bbsim.AlarmType_Types{
	"DyingGasp":                bbsim.AlarmType_DYING_GASP,
	"StartupFailure":           bbsim.AlarmType_ONU_STARTUP_FAILURE,
	"SignalDegrade":            bbsim.AlarmType_ONU_SIGNAL_DEGRADE,
	"DriftOfWindow":            bbsim.AlarmType_ONU_DRIFT_OF_WINDOW,
	"LossOfOmciChannel":        bbsim.AlarmType_ONU_LOSS_OF_OMCI_CHANNEL,
	"SignalsFailure":           bbsim.AlarmType_ONU_SIGNALS_FAILURE,
	"TransmissionInterference": bbsim.AlarmType_ONU_TRANSMISSION_INTERFERENCE_WARNING,
	"ActivationFailure":        bbsim.AlarmType_ONU_ACTIVATION_FAILURE,
	"ProcessingError":          bbsim.AlarmType_ONU_PROCESSING_ERROR,
	"LossOfKeySyncFailure":     bbsim.AlarmType_ONU_LOSS_OF_KEY_SYNC_FAILURE,

	// Break out OnuAlarm into its subcases.
	"LossOfSignal":   bbsim.AlarmType_ONU_ALARM_LOS,
	"LossOfBurst":    bbsim.AlarmType_ONU_ALARM_LOB,
	"LOPC_MISS":      bbsim.AlarmType_ONU_ALARM_LOPC_MISS,
	"LOPC_MIC_ERROR": bbsim.AlarmType_ONU_ALARM_LOPC_MIC_ERROR,
	"LossOfFrame":    bbsim.AlarmType_ONU_ALARM_LOFI,
	"LossOfPloam":    bbsim.AlarmType_ONU_ALARM_LOAMI,

	// Whole-PON / Non-onu-specific
	"PonLossOfSignal": bbsim.AlarmType_LOS,
}

func AlarmNameToEnum(name string) (*bbsim.AlarmType_Types, error) {
	v, okay := AlarmNameMap[name]
	if !okay {
		return nil, fmt.Errorf("Unknown Alarm Name: %v", name)
	}

	return &v, nil
}

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

// BuildAlarmIndication function forms openolt alarmIndication as per AlarmRequest
func BuildAlarmIndication(req *bbsim.AlarmRequest, o *devices.OltDevice) (*openolt.AlarmIndication, error) {
	var alarm *openolt.AlarmIndication
	var onu *devices.Onu
	var err error

	alarmType, err := AlarmNameToEnum(req.AlarmType)
	if err != nil {
		return nil, err
	}

	if *alarmType != bbsim.AlarmType_LOS {
		// No ONU Id for LOS
		onu, err = o.FindOnuBySn(req.SerialNumber)
		if err != nil {
			return nil, err
		}
	}

	switch *alarmType {
	case bbsim.AlarmType_LOS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_LosInd{&openolt.LosIndication{
				// No ONU Id for LOS
				Status: req.Status,
				IntfId: uint32(extractInt(req.Parameters, "InterfaceId", 0)),
			}},
		}
	case bbsim.AlarmType_DYING_GASP:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_DyingGaspInd{&openolt.DyingGaspIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_STARTUP_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuStartupFailInd{&openolt.OnuStartupFailureIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_SIGNAL_DEGRADE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuSignalDegradeInd{&openolt.OnuSignalDegradeIndication{
				Status:              req.Status,
				OnuId:               onu.ID,
				IntfId:              onu.PonPortID,
				InverseBitErrorRate: uint32(extractInt(req.Parameters, "InverseBitErrorRate", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_SIGNALS_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuSignalsFailInd{&openolt.OnuSignalsFailureIndication{
				Status:              req.Status,
				OnuId:               onu.ID,
				IntfId:              onu.PonPortID,
				InverseBitErrorRate: uint32(extractInt(req.Parameters, "InverseBitErrorRate", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_DRIFT_OF_WINDOW:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuDriftOfWindowInd{&openolt.OnuDriftOfWindowIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
				Drift:  uint32(extractInt(req.Parameters, "Drift", 0)),
				NewEqd: uint32(extractInt(req.Parameters, "NewEqd", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_LOSS_OF_OMCI_CHANNEL:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuLossOmciInd{&openolt.OnuLossOfOmciChannelIndication{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_TRANSMISSION_INTERFERENCE_WARNING:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuTiwiInd{&openolt.OnuTransmissionInterferenceWarning{
				Status: req.Status,
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
				Drift:  uint32(extractInt(req.Parameters, "Drift", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_ACTIVATION_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuActivationFailInd{&openolt.OnuActivationFailureIndication{
				OnuId:      onu.ID,
				IntfId:     onu.PonPortID,
				FailReason: uint32(extractInt(req.Parameters, "FailReason", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_PROCESSING_ERROR:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuProcessingErrorInd{&openolt.OnuProcessingErrorIndication{
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_LOSS_OF_KEY_SYNC_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuLossOfSyncFailInd{&openolt.OnuLossOfKeySyncFailureIndication{
				OnuId:  onu.ID,
				IntfId: onu.PonPortID,
				Status: req.Status,
			}},
		}
	case bbsim.AlarmType_ONU_ITU_PON_STATS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuItuPonStatsInd{&openolt.OnuItuPonStatsIndication{
				OnuId:     onu.ID,
				IntfId:    onu.PonPortID,
				RdiErrors: uint32(extractInt(req.Parameters, "RdiErrors", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LosStatus: req.Status,
				OnuId:     onu.ID,
				IntfId:    onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOB:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LobStatus: req.Status,
				OnuId:     onu.ID,
				IntfId:    onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOPC_MISS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LopcMissStatus: req.Status,
				OnuId:          onu.ID,
				IntfId:         onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOPC_MIC_ERROR:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LopcMicErrorStatus: req.Status,
				OnuId:              onu.ID,
				IntfId:             onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOFI:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LofiStatus: req.Status,
				OnuId:      onu.ID,
				IntfId:     onu.PonPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOAMI:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{&openolt.OnuAlarmIndication{
				LoamiStatus: req.Status,
				OnuId:       onu.ID,
				IntfId:      onu.PonPortID,
			}},
		}
	default:
		return nil, fmt.Errorf("Unknown alarm type %v", req.AlarmType)
	}

	return alarm, nil
}

// SimulateAlarm accept request for alarms and send proper alarmIndication to openolt stream
func SimulateAlarm(ctx context.Context, req *bbsim.AlarmRequest, o *devices.OltDevice) error {
	alarmIndication, err := BuildAlarmIndication(req, o)
	if err != nil {
		return err
	}

	err = o.SendAlarmIndication(ctx, alarmIndication)
	if err != nil {
		return err
	}

	return nil
}
