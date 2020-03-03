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
	"strconv"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OnuAlarmNameMap string to enum map
var OnuAlarmNameMap = map[string]bbsim.AlarmType_Types{
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
}

// OltAlarmNameMap string to enum map
var OltAlarmNameMap = map[string]bbsim.AlarmType_Types{
	// PON / Non-onu-specific
	"PonLossOfSignal": bbsim.AlarmType_LOS,
	// NNI / Non-onu-specific
	"NniLossOfSignal": bbsim.AlarmType_LOS,
}

func AlarmNameToEnum(name string) (*bbsim.AlarmType_Types, error) {
	v, okay := OnuAlarmNameMap[name]
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

// BuildOnuAlarmIndication function forms openolt alarmIndication as per ONUAlarmRequest
func BuildOnuAlarmIndication(req *bbsim.ONUAlarmRequest, o *devices.OltDevice) (*openolt.AlarmIndication, error) {
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

// SimulateOnuAlarm accept request for Onu alarms and send proper alarmIndication to openolt stream
func SimulateOnuAlarm(ctx context.Context, req *bbsim.ONUAlarmRequest, o *devices.OltDevice) error {
	alarmIndication, err := BuildOnuAlarmIndication(req, o)
	if err != nil {
		return err
	}

	err = o.SendAlarmIndication(ctx, alarmIndication)
	if err != nil {
		return err
	}

	return nil
}

// InterfaceIDToPortNo converts InterfaceID to voltha PortID
// Refer openolt adapter code(master) voltha-openolt-adapter/adaptercore/olt_platform.go: IntfIDToPortNo()
func InterfaceIDToPortNo(req *bbsim.OLTAlarmRequest) uint32 {
	// Converts interface-id to port-numbers that can be understood by the VOLTHA
	if req.InterfaceType == "nni" || req.InterfaceType == "PonLossOfSignal" {
		// nni at voltha starts with 1,048,576
		// nni = 1,048,576 + InterfaceID
		return 0x1<<20 + req.InterfaceID
	} else if req.InterfaceType == "pon" || req.InterfaceType == "NniLossOfSignal" {
		// pon = 536,870,912 + InterfaceID
		return (0x2 << 28) + req.InterfaceID
		// In bbsim, pon starts from 1
	}
	return 0
}

// IsPonPortPresentInOlt verifies if given Pon port is present in olt
func IsPonPortPresentInOlt(PonPort uint32) bool {
	o := devices.GetOLT()
	for _, intf := range o.Pons {
		if intf.ID == PonPort {
			return true
		}
	}
	return false
}

// IsNniPortPresentInOlt verifies if given nni port is present in olt
func IsNniPortPresentInOlt(nniPort uint32) bool {
	o := devices.GetOLT()
	for _, intf := range o.Nnis {
		if intf.ID == nniPort {
			return true
		}
	}
	return false
}

// SimulateOltAlarm accept request for Olt alarms and send proper alarmIndication to openolt stream
func SimulateOltAlarm(ctx context.Context, req *bbsim.OLTAlarmRequest, o *devices.OltDevice) error {
	var alarmIndication *openolt.AlarmIndication
	var err error

	//check if its a valid port id
	switch req.InterfaceType {
	case "nni":
		if !IsNniPortPresentInOlt(uint32(req.InterfaceID)) {
			return status.Errorf(codes.NotFound, strconv.Itoa(int(req.InterfaceID))+" NNI not present in olt")
		}

	case "pon":
		if !IsPonPortPresentInOlt(uint32(req.InterfaceID)) {
			return status.Errorf(codes.NotFound, strconv.Itoa(int(req.InterfaceID))+" PON not present in olt")
		}
	}
	alarmIndication = &openolt.AlarmIndication{
		Data: &openolt.AlarmIndication_LosInd{&openolt.LosIndication{
			Status: req.Status,
			IntfId: InterfaceIDToPortNo(req),
		}},
	}

	err = o.SendAlarmIndication(ctx, alarmIndication)
	if err != nil {
		return err
	}

	return nil
}
