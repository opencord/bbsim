/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"fmt"
	"github.com/opencord/bbsim/internal/bbsim/types"
	"strconv"

	"github.com/opencord/bbsim/internal/common"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/voltha-protos/v5/go/openolt"
)

func AlarmNameToEnum(name string) (bbsim.AlarmType_Types, error) {
	v, okay := common.ONUAlarms[name]
	if !okay {
		return 0, fmt.Errorf("Unknown Alarm Name: %v", name)
	}

	return v, nil
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

// BuildOnuAlarmIndication function forms openolt alarmIndication as per ONUAlarmRequest
func BuildOnuAlarmIndication(req *bbsim.ONUAlarmRequest, onuID uint32, ponPortID uint32) (*openolt.AlarmIndication, error) {
	var alarm *openolt.AlarmIndication
	var err error

	alarmType, err := AlarmNameToEnum(req.AlarmType)
	if err != nil {
		return nil, err
	}

	switch alarmType {
	case bbsim.AlarmType_DYING_GASP:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_DyingGaspInd{DyingGaspInd: &openolt.DyingGaspIndication{
				Status: req.Status,
				OnuId:  onuID,
				IntfId: ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_STARTUP_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuStartupFailInd{OnuStartupFailInd: &openolt.OnuStartupFailureIndication{
				Status: req.Status,
				OnuId:  onuID,
				IntfId: ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_SIGNAL_DEGRADE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuSignalDegradeInd{OnuSignalDegradeInd: &openolt.OnuSignalDegradeIndication{
				Status:              req.Status,
				OnuId:               onuID,
				IntfId:              ponPortID,
				InverseBitErrorRate: uint32(extractInt(req.Parameters, "InverseBitErrorRate", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_SIGNALS_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuSignalsFailInd{OnuSignalsFailInd: &openolt.OnuSignalsFailureIndication{
				Status:              req.Status,
				OnuId:               onuID,
				IntfId:              ponPortID,
				InverseBitErrorRate: uint32(extractInt(req.Parameters, "InverseBitErrorRate", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_DRIFT_OF_WINDOW:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuDriftOfWindowInd{OnuDriftOfWindowInd: &openolt.OnuDriftOfWindowIndication{
				Status: req.Status,
				OnuId:  onuID,
				IntfId: ponPortID,
				Drift:  uint32(extractInt(req.Parameters, "Drift", 0)),
				NewEqd: uint32(extractInt(req.Parameters, "NewEqd", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_LOSS_OF_OMCI_CHANNEL:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuLossOmciInd{OnuLossOmciInd: &openolt.OnuLossOfOmciChannelIndication{
				Status: req.Status,
				OnuId:  onuID,
				IntfId: ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_TRANSMISSION_INTERFERENCE_WARNING:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuTiwiInd{OnuTiwiInd: &openolt.OnuTransmissionInterferenceWarning{
				Status: req.Status,
				OnuId:  onuID,
				IntfId: ponPortID,
				Drift:  uint32(extractInt(req.Parameters, "Drift", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_ACTIVATION_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuActivationFailInd{OnuActivationFailInd: &openolt.OnuActivationFailureIndication{
				OnuId:      onuID,
				IntfId:     ponPortID,
				FailReason: uint32(extractInt(req.Parameters, "FailReason", 0)),
			}},
		}
	case bbsim.AlarmType_ONU_PROCESSING_ERROR:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuProcessingErrorInd{OnuProcessingErrorInd: &openolt.OnuProcessingErrorIndication{
				OnuId:  onuID,
				IntfId: ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_LOSS_OF_KEY_SYNC_FAILURE:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuLossOfSyncFailInd{OnuLossOfSyncFailInd: &openolt.OnuLossOfKeySyncFailureIndication{
				OnuId:  onuID,
				IntfId: ponPortID,
				Status: req.Status,
			}},
		}
	case bbsim.AlarmType_ONU_ITU_PON_STATS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuItuPonStatsInd{OnuItuPonStatsInd: &openolt.OnuItuPonStatsIndication{
				OnuId:  onuID,
				IntfId: ponPortID,
				Stats: &openolt.OnuItuPonStatsIndication_RdiErrorInd{
					RdiErrorInd: &openolt.RdiErrorIndication{
						RdiErrorCount: uint64(extractInt(req.Parameters, "RdiErrors", 0)),
						Status:        req.Status,
					},
				},
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LosStatus: req.Status,
				OnuId:     onuID,
				IntfId:    ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOB:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LobStatus: req.Status,
				OnuId:     onuID,
				IntfId:    ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOPC_MISS:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LopcMissStatus: req.Status,
				OnuId:          onuID,
				IntfId:         ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOPC_MIC_ERROR:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LopcMicErrorStatus: req.Status,
				OnuId:              onuID,
				IntfId:             ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOFI:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LofiStatus: req.Status,
				OnuId:      onuID,
				IntfId:     ponPortID,
			}},
		}
	case bbsim.AlarmType_ONU_ALARM_LOAMI:
		alarm = &openolt.AlarmIndication{
			Data: &openolt.AlarmIndication_OnuAlarmInd{OnuAlarmInd: &openolt.OnuAlarmIndication{
				LoamiStatus: req.Status,
				OnuId:       onuID,
				IntfId:      ponPortID,
			}},
		}
	default:
		return nil, fmt.Errorf("Unknown ONU alarm type %v", req.AlarmType)
	}

	return alarm, nil
}

// SimulateOnuAlarm accept request for Onu alarms and send proper alarmIndication to openolt stream
func SimulateOnuAlarm(req *bbsim.ONUAlarmRequest, onuID uint32, ponPortID uint32, channel chan types.Message) error {
	alarmIndication, err := BuildOnuAlarmIndication(req, onuID, ponPortID)
	if err != nil {
		return err
	}

	msg := types.Message{
		Type: types.AlarmIndication,
		Data: alarmIndication,
	}

	channel <- msg

	return nil
}
