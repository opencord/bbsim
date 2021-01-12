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
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	log "github.com/sirupsen/logrus"
)

func ParseSetRequest(omciPkt gopacket.Packet) (*omci.SetRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeSetRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeSetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.SetRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeSetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateSetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseSetRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  msgObj.AttributeMask,
	}).Debug("received-omci-set-request")

	response := &omci.SetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: 0,
		},
		Result: me.Success,
	}

	pkt, err := serialize(omci.SetResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot serialize SetResponse")
		return nil, err
	}

	return pkt, nil

}

// TODO find a better way to integrate this with the ONU state machine
func SendUniStatusAlarm(adminState uint8, ponPortId uint32, onuId uint32, onuSn string, stream openolt.Openolt_EnableIndicationServer) {

	// TODO generate using omci-lib-go
	linkMsgDown := []byte{
		0x00, 0x00, 0x10, 0x0a, 0x00, 0x0b, 0x01, 0x01,
		0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	linkMsgUp := []byte{
		0x00, 0x00, 0x10, 0x0a, 0x00, 0x0b, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	var pkt []byte
	if adminState == 0 {
		pkt = linkMsgUp
	} else if adminState == 1 {
		pkt = linkMsgDown
	}

	indication := &openolt.Indication_OmciInd{
		OmciInd: &openolt.OmciIndication{
			IntfId: ponPortId,
			OnuId:  onuId,
			Pkt:    pkt,
		},
	}
	if err := stream.Send(&openolt.Indication{Data: indication}); err != nil {
		omciLogger.WithFields(log.Fields{
			"IntfId":       ponPortId,
			"SerialNumber": onuSn,
			"omciPacket":   indication.OmciInd.Pkt,
		}).Errorf("Failed to send UNI Link Alarm: %v", err)
		return
	}
	omciLogger.WithFields(log.Fields{
		"IntfId":       ponPortId,
		"SerialNumber": onuSn,
		"omciPacket":   indication.OmciInd.Pkt,
	}).Debug("UNI Link alarm sent")
}
