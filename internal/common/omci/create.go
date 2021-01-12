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
	"encoding/hex"
	"errors"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ParseCreateRequest(omciPkt gopacket.Packet) (*omci.CreateRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeCreateRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeCreateRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.CreateRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeCreateRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateCreateResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseCreateRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
	}).Trace("recevied-omci-create-request")

	response := &omci.CreateResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		Result: me.Success,
	}

	pkt, err := serialize(omci.CreateResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-serialize-CreateResponse")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Trace("omci-create-response")

	return pkt, nil
}

// methods used by BBR to drive the OMCI state machine

func CreateGalEnetRequest(tid uint16) ([]byte, error) {
	params := me.ParamData{
		EntityID:   galEthernetEID,
		Attributes: me.AttributeValueMap{"MaximumGemPayloadSize": maxGemPayloadSize},
	}
	meDef, _ := me.NewGalEthernetProfile(params)
	pkt, err := omci.GenFrame(meDef, omci.CreateRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate GalEnetRequest")
	}
	return hexEncode(pkt)
}

func CreateEnableUniRequest(tid uint16, uniId uint16, enabled bool, isPtp bool) ([]byte, error) {

	var _enabled uint8
	if enabled {
		_enabled = uint8(1)
	} else {
		_enabled = uint8(0)
	}

	data := me.ParamData{
		EntityID: uniId,
		Attributes: me.AttributeValueMap{
			"AdministrativeState": _enabled,
		},
	}
	var medef *me.ManagedEntity
	var omciErr me.OmciErrors

	if isPtp {
		medef, omciErr = me.NewPhysicalPathTerminationPointEthernetUni(data)
	} else {
		medef, omciErr = me.NewVirtualEthernetInterfacePoint(data)
	}
	if omciErr != nil {
		return nil, omciErr.GetError()
	}
	pkt, err := omci.GenFrame(medef, omci.SetRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate EnableUniRequest")
	}
	return hexEncode(pkt)
}

func CreateGemPortRequest(tid uint16) ([]byte, error) {
	params := me.ParamData{
		EntityID: gemEID,
		Attributes: me.AttributeValueMap{
			"PortId":                              1,
			"TContPointer":                        1,
			"Direction":                           0,
			"TrafficManagementPointerForUpstream": 0,
			"TrafficDescriptorProfilePointerForUpstream": 0,
			"UniCounter":                                   0,
			"PriorityQueuePointerForDownStream":            0,
			"EncryptionState":                              0,
			"TrafficDescriptorProfilePointerForDownstream": 0,
			"EncryptionKeyRing":                            0,
		},
	}
	meDef, _ := me.NewGemPortNetworkCtp(params)
	pkt, err := omci.GenFrame(meDef, omci.CreateRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate GemPortRequest")
	}
	return hexEncode(pkt)
}
