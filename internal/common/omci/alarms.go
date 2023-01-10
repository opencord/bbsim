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

package omci

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

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

func CreateGetAllAlarmsResponse(omciMsg *omci.OMCI, onuAlarmDetails map[OnuAlarmInfoMapKey]OnuAlarmInfo) ([]byte, error) {
	var alarmEntityClass me.ClassID
	var noOfCommands uint16 = 0
	var isExtended bool = false

	if omciMsg.DeviceIdentifier == omci.ExtendedIdent {
		isExtended = true
	}
	alarmEntityClass = me.PhysicalPathTerminationPointEthernetUniClassID //Currently doing it for PPTP classID only

	key := OnuAlarmInfoMapKey{
		MeClassID: alarmEntityClass,
	}
	for i := 257; i < 261; i++ { //Currently doing it for up to four PPTP instances only
		key.MeInstance = uint16(i)
		if _, ok := onuAlarmDetails[key]; ok {
			noOfCommands++
		}
	}
	response := &omci.GetAllAlarmsResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
			Extended:    isExtended,
		},
		NumberOfCommands: noOfCommands,
	}
	omciLayer := &omci.OMCI{
		TransactionID:    omciMsg.TransactionID,
		MessageType:      omci.GetAllAlarmsResponseType,
		DeviceIdentifier: omciMsg.DeviceIdentifier,
	}
	var options gopacket.SerializeOptions
	options.FixLengths = true

	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, options, omciLayer, response)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err":  err,
			"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		}).Error("cannot-Serialize-GetAllAlarmsResponse")
		return nil, err
	}
	pkt := buffer.Bytes()

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Trace("omci-get-all-alarms-response")

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
	if msgObj.CommandSequenceNumber > 4 { //Currently doing it for up to four PPTP instances only
		omciLogger.Warn("unsupported-CommandSequenceNumber-in-get-all-alarm-next", msgObj.CommandSequenceNumber)
		return nil, fmt.Errorf("unspported-command-sequence-number-in-get-all-alarms-next")
	}
	var alarmEntityClass me.ClassID
	var meInstance uint16
	var alarmBitMap [28]byte
	var isExtended bool
	var additionalAlarms omci.AdditionalAlarmsData

	if omciMsg.DeviceIdentifier == omci.ExtendedIdent {
		isExtended = true
	} else {
		isExtended = false
	}
	alarmEntityClass = me.PhysicalPathTerminationPointEthernetUniClassID //Currently doing it for PPTP classID only

	key := OnuAlarmInfoMapKey{
		MeClassID: alarmEntityClass,
	}
	meInstance = 257 + msgObj.CommandSequenceNumber
	key.MeInstance = meInstance
	if alarmInfo, ok := onuAlarmDetails[key]; ok {
		alarmBitMap = alarmInfo.AlarmBitMap
	} else {
		return nil, fmt.Errorf("alarm-info-for-me-not-present-in-alarm-info-map")
	}
	response := &omci.GetAllAlarmsNextResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
			Extended:    isExtended,
		},
		AlarmEntityClass:    alarmEntityClass,
		AlarmEntityInstance: meInstance,
		AlarmBitMap:         alarmBitMap,
	}
	if isExtended && msgObj.CommandSequenceNumber == 0 {
		for i := 258; i < 261; i++ {
			key.MeInstance = uint16(i)
			if addAlarmInfo, ok := onuAlarmDetails[key]; ok {
				additionalAlarms.AlarmEntityClass = key.MeClassID
				additionalAlarms.AlarmEntityInstance = uint16(i)
				additionalAlarms.AlarmBitMap = addAlarmInfo.AlarmBitMap
				response.AdditionalAlarms = append(response.AdditionalAlarms, additionalAlarms)
			}
		}
	}
	omciLayer := &omci.OMCI{
		TransactionID:    omciMsg.TransactionID,
		MessageType:      omci.GetAllAlarmsNextResponseType,
		DeviceIdentifier: omciMsg.DeviceIdentifier,
	}
	var options gopacket.SerializeOptions
	options.FixLengths = true

	buffer := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(buffer, options, omciLayer, response)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err":  err,
			"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		}).Error("cannot-Serialize-GetAllAlarmsNextResponse")
		return nil, err
	}
	pkt := buffer.Bytes()

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Trace("omci-get-all-alarms-next-response")
	return pkt, nil
}
