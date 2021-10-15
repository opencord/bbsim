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

package omci

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	log "github.com/sirupsen/logrus"
)

type OnuAlarmInfo struct {
	SequenceNo  uint8
	AlarmBitMap [28]byte
}
type OnuAlarmInfoMapKey struct {
	MeInstance uint16
	MeClassID  me.ClassID
}

// CreateUniStatusAlarm will generate an Alarm packet to report that the Link is UP or DOWN
// as a consequence of a SetRequest on PhysicalPathTerminationPointEthernetUniClassID
func CreateUniStatusAlarm(raiseAlarm bool, entityId uint16, sequenceNo uint8) ([]byte, [28]byte) {
	notif := &omci.AlarmNotificationMsg{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.PhysicalPathTerminationPointEthernetUniClassID,
			EntityInstance: entityId,
		},
		AlarmSequenceNumber: byte(sequenceNo),
	}
	if raiseAlarm {
		//PPTP class has only 0th bit alarm for UNI LOS
		if err := notif.ActivateAlarm(0); err != nil {
			omciLogger.WithFields(log.Fields{
				"Err": err,
			}).Error("Cannot Create AlarmNotificationMsg")
			return nil, [28]byte{}
		}
	} else {
		if err := notif.ClearAlarm(0); err != nil {
			omciLogger.WithFields(log.Fields{
				"Err": err,
			}).Error("Cannot Create AlarmNotificationMsg")
			return nil, [28]byte{}
		}
	}
	pkt, err := Serialize(omci.AlarmNotificationType, notif, 0)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot Serialize AlarmNotificationMsg")
		return nil, [28]byte{}
	}
	return pkt, notif.AlarmBitmap
}

func CreateGetAllAlarmsResponse(tid uint16, onuAlarmDetails map[OnuAlarmInfoMapKey]OnuAlarmInfo) ([]byte, error) {
	var alarmEntityClass me.ClassID
	var meInstance uint16
	var noOfCommands uint16 = 0
	alarmEntityClass = me.PhysicalPathTerminationPointEthernetUniClassID //Currently doing for PPTP classID
	meInstance = 257
	key := OnuAlarmInfoMapKey{
		MeInstance: meInstance,
		MeClassID:  alarmEntityClass,
	}
	if _, ok := onuAlarmDetails[key]; ok {
		noOfCommands = 1
	}
	numberOfCommands := uint16(noOfCommands)

	request := &omci.GetAllAlarmsResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		NumberOfCommands: numberOfCommands,
	}
	pkt, err := Serialize(omci.GetAllAlarmsResponseType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot Serialize GetAllAlarmsResponse")
		return nil, err
	}
	return pkt, nil
}
func ParseGetAllAlarmsNextRequest(omciPkt gopacket.Packet) (*omci.GetAllAlarmsNextRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeGetAllAlarmsNextRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeGetAllAlarmsNextRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.GetAllAlarmsNextRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for GetAllAlarmsNextRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateGetAllAlarmsNextResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI, onuAlarmDetails map[OnuAlarmInfoMapKey]OnuAlarmInfo) ([]byte, error) {

	msgObj, err := ParseGetAllAlarmsNextRequest(omciPkt)
	if err != nil {
		err := "omci Msg layer could not be assigned for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":           msgObj.EntityClass,
		"EntityInstance":        msgObj.EntityInstance,
		"CommandSequenceNumber": msgObj.CommandSequenceNumber,
	}).Trace("received-omci-get-all-alarms-next-request")

	var alarmEntityClass me.ClassID
	var meInstance uint16
	var alarmBitMap [28]byte

	switch msgObj.CommandSequenceNumber {
	case 0:
		alarmEntityClass = me.PhysicalPathTerminationPointEthernetUniClassID
		meInstance = 257
		//Checking if the alarm is raised in the bitmap, we will send clear just to generate missed clear alarm and
		// vice versa.
		key := OnuAlarmInfoMapKey{
			MeInstance: meInstance,
			MeClassID:  alarmEntityClass,
		}
		if alarmInfo, ok := onuAlarmDetails[key]; ok {
			alarmBitMap = alarmInfo.AlarmBitMap
		} else {
			return nil, fmt.Errorf("alarm-info-for-me-not-present-in-alarm-info-map")
		}
	default:
		omciLogger.Warn("unsupported-CommandSequenceNumber-in-get-all-alarm-next", msgObj.CommandSequenceNumber)
		return nil, fmt.Errorf("unspported-command-sequence-number-in-get-all-alarms-next")
	}

	response := &omci.GetAllAlarmsNextResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		AlarmEntityClass:    alarmEntityClass,
		AlarmEntityInstance: meInstance,
		AlarmBitMap:         alarmBitMap,
	}

	omciLogger.WithFields(log.Fields{
		"AlarmEntityClass":    alarmEntityClass,
		"AlarmEntityInstance": meInstance,
		"AlarmBitMap":         alarmBitMap,
	}).Trace("created-omci-getAllAlarmsNext-response")

	pkt, err := Serialize(omci.GetAllAlarmsNextResponseType, response, omciMsg.TransactionID)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize GetAllAlarmsNextRequest")
		return nil, err
	}

	return pkt, nil
}
