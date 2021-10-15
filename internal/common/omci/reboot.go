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

func ParseRebootRequest(omciPkt gopacket.Packet) (*omci.RebootRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeRebootRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeRebootRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.RebootRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeRebootRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateRebootResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseRebootRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
	}).Trace("received-omci-set-request")

	response := &omci.SetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		Result: me.Success,
	}

	pkt, err := Serialize(omci.RebootResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-Serialize-RebootResponse")
		return nil, err
	}

	return pkt, nil

}
