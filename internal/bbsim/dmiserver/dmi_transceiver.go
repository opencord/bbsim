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

package dmiserver

import (
	"context"
	"fmt"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	dmi "github.com/opencord/device-management-interface/go/dmi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ponInterfaceType = "pon"
	alarmStatusRaise = "on"
	alarmStatusClear = "off"
)

type Transceiver struct {
	ID         uint32
	Uuid       string
	Name       string
	Pons       []*devices.PonPort
	Technology dmi.TransceiverType
	//Setting this bool will prevent the transceiver
	//from being plugged out while already out, and
	//plugged in while already in, but won't prevent
	//the associated PONs from being enabled in other
	//ways while the transceiver is plugged out
	PluggedIn bool
}

func newTransceiver(id uint32, pons []*devices.PonPort) *Transceiver {
	return &Transceiver{
		ID:         id,
		Uuid:       getTransceiverUUID(id).Uuid,
		Name:       getTransceiverName(id),
		Pons:       pons,
		Technology: dmi.TransceiverType_TYPE_UNDEFINED,
		PluggedIn:  true,
	}
}

func getTransceiverWithId(transId uint32, dms *DmiAPIServer) (*Transceiver, error) {
	for _, t := range dms.Transceivers {
		if t.ID == transId {
			return t, nil
		}
	}

	return nil, fmt.Errorf("Cannot find transceiver with ID %d", transId)
}

/////// Handler methods for grpc API

func (s DmiAPIServer) GetTransceivers(ctx context.Context, req *bbsim.DmiEmpty) (*bbsim.Transceivers, error) {
	res := &bbsim.Transceivers{
		Items: []*bbsim.Transceiver{},
	}

	for _, t := range s.Transceivers {
		item := bbsim.Transceiver{
			ID:         t.ID,
			UUID:       t.Uuid,
			Name:       t.Name,
			Technology: t.Technology.String(),
			PluggedIn:  t.PluggedIn,
			PonIds:     []uint32{},
		}

		for _, pon := range t.Pons {
			item.PonIds = append(item.PonIds, pon.ID)
		}

		res.Items = append(res.Items, &item)
	}

	return res, nil
}

// PlugOutTransceiver plugs out the transceiver by its ID
func (s DmiAPIServer) PlugOutTransceiver(ctx context.Context, req *bbsim.TransceiverRequest) (*bbsim.DmiResponse, error) {
	logger.WithFields(log.Fields{
		"IntfId": req.TransceiverId,
	}).Infof("Received request to plug out PON transceiver")

	res := &bbsim.DmiResponse{}
	olt := devices.GetOLT()

	//Generate DMI event
	dmiServ, err := getDmiAPIServer()
	if err != nil {
		res.StatusCode = int32(codes.Unavailable)
		res.Message = fmt.Sprintf("Cannot get DMI server instance: %v", err)
		return res, nil
	}

	trans, err := getTransceiverWithId(req.TransceiverId, dmiServ)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = fmt.Sprintf("Cannot find transceiver with ID %d: %v", req.TransceiverId, err)
		return res, nil
	}

	if !trans.PluggedIn {
		res.StatusCode = int32(codes.Aborted)
		res.Message = fmt.Sprintf("Cannot plug out transceiver with ID %d since it's not plugged in", req.TransceiverId)
		return res, nil
	}

	err = PlugoutTransceiverComponent(req.TransceiverId, dmiServ)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = fmt.Sprintf("Cannot remove transceiver with ID %d: %v", req.TransceiverId, err)
		return res, nil
	}
	logger.Debug("Removed transceiver from DMI inventory")

	if olt.InternalState.Is(devices.OltInternalStateEnabled) {
		logger.Debug("Sending alarms for transceiver plug out")
		for _, pon := range trans.Pons {
			if pon.InternalState.Is("enabled") {

				if err = olt.SetAlarm(pon.ID, ponInterfaceType, alarmStatusRaise); err != nil {
					logger.WithFields(log.Fields{
						"ponId": pon.ID,
						"err":   err,
					}).Error("Cannot raise LOS alarm for PON")
				}

				if err = pon.InternalState.Event("disable"); err != nil {
					logger.WithFields(log.Fields{
						"ponId": pon.ID,
						"err":   err,
					}).Error("Cannot disable PON")
					continue
				}

				for _, onu := range pon.Onus {
					if err := onu.SetAlarm(bbsim.AlarmType_ONU_ALARM_LOS.String(), alarmStatusRaise); err != nil {
						logger.WithFields(log.Fields{
							"ponId": pon.ID,
							"onuId": onu.ID,
							"err":   err,
						}).Error("Cannot raise LOS alarm for ONU")
					}
				}
			}
		}
	} else {
		logger.Debug("No operation on devices since the OLT is not enabled")
	}

	event := dmi.Event{
		EventId: dmi.EventIds_EVENT_TRANSCEIVER_PLUG_OUT,
		EventMetadata: &dmi.EventMetaData{
			DeviceUuid: dmiServ.uuid,
			ComponentUuid: &dmi.Uuid{
				Uuid: trans.Uuid,
			},
			ComponentName: trans.Name,
		},
		RaisedTs: timestamppb.Now(),
	}

	sendOutEventOnKafka(event, dmiServ)
	logger.Debug("Transceiver plug out event sent")

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Plugged out transceiver %d", req.TransceiverId)

	return res, nil
}

// PlugInTransceiver plugs in the transceiver by its ID
func (s DmiAPIServer) PlugInTransceiver(ctx context.Context, req *bbsim.TransceiverRequest) (*bbsim.DmiResponse, error) {
	logger.WithFields(log.Fields{
		"IntfId": req.TransceiverId,
	}).Infof("Received request to plug in PON transceiver")

	res := &bbsim.DmiResponse{}
	olt := devices.GetOLT()

	//Generate DMI event
	dmiServ, err := getDmiAPIServer()
	if err != nil {
		res.StatusCode = int32(codes.Unavailable)
		res.Message = fmt.Sprintf("Cannot get DMI server instance: %v", err)
		return res, nil
	}

	trans, err := getTransceiverWithId(req.TransceiverId, dmiServ)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = fmt.Sprintf("Cannot find transceiver with ID %d: %v", req.TransceiverId, err)
		return res, nil
	}

	if trans.PluggedIn {
		res.StatusCode = int32(codes.Aborted)
		res.Message = fmt.Sprintf("Cannot plug in transceiver with ID %d since it's already plugged in", req.TransceiverId)
		return res, nil
	}

	err = PluginTransceiverComponent(req.TransceiverId, dmiServ)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = fmt.Sprintf("Cannot add transceiver with ID %d: %v", req.TransceiverId, err)
		return res, nil
	}
	logger.Debug("Added transceiver to DMI inventory")

	if olt.InternalState.Is(devices.OltInternalStateEnabled) {
		logger.Debug("Sending alarms for transceiver plug in")
		for _, pon := range trans.Pons {

			if err = olt.SetAlarm(pon.ID, ponInterfaceType, alarmStatusClear); err != nil {
				logger.WithFields(log.Fields{
					"ponId": pon.ID,
					"err":   err,
				}).Error("Cannot clear LOS alarm for ONU")
			}

			if err = pon.InternalState.Event("enable"); err != nil {
				logger.WithFields(log.Fields{
					"ponId": pon.ID,
					"err":   err,
				}).Error("Cannot enable PON")
				continue
			}

			for _, onu := range pon.Onus {
				if err := onu.SetAlarm(bbsim.AlarmType_ONU_ALARM_LOS.String(), alarmStatusClear); err != nil {
					logger.WithFields(log.Fields{
						"ponId": pon.ID,
						"onuId": onu.ID,
						"err":   err,
					}).Error("Cannot clear LOS alarm for ONU")
				}
			}
		}
	} else {
		logger.Debug("No operation on devices since the OLT is not enabled")
	}

	event := dmi.Event{
		EventId: dmi.EventIds_EVENT_TRANSCEIVER_PLUG_IN,
		EventMetadata: &dmi.EventMetaData{
			DeviceUuid: dmiServ.uuid,
			ComponentUuid: &dmi.Uuid{
				Uuid: trans.Uuid,
			},
			ComponentName: trans.Name,
		},
		RaisedTs: timestamppb.Now(),
	}

	sendOutEventOnKafka(event, dmiServ)
	logger.Debug("Transceiver plug in event sent")

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Plugged in transceiver %d", req.TransceiverId)

	return res, nil
}
