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
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
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

func CreateSetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI, result me.Results) ([]byte, error) {

	msgObj, err := ParseSetRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  msgObj.AttributeMask,
	}).Trace("received-omci-set-request")

	response := &omci.SetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		Result: result,
	}

	pkt, err := Serialize(omci.SetResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-Serialize-SetResponse")
		return nil, err
	}

	return pkt, nil

}

func CreateSetRequest(managedEntity *me.ManagedEntity, tid uint16) ([]byte, error) {

	// TODO
	// why can't we create the SetRequest as we do for all other omci Requests (eg: MibResetRequest)?
	// if we do the Attributes are not sent

	request := &omci.SetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    managedEntity.GetClassID(),
			EntityInstance: managedEntity.GetEntityID(),
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: 0x800, // FIXME how can we generate this based on managedEntity.AttributeValueMap?
	}
	omciLogger.Info(request)

	pkt, err := Serialize(omci.SetRequestType, request, tid)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize SetRequest")
		return nil, err
	}
	return HexEncode(pkt)
}
