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
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ParseGetRequest(omciPkt gopacket.Packet) (*omci.GetRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeGetRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.GetRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateGetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseGetRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  msgObj.AttributeMask,
	}).Debug("recevied-omci-get-request")

	switch msgObj.EntityClass {
	case me.Onu2GClassID:
		return CreateOnu2gResponse(omciMsg.TransactionID)
	default:
		log.Warnf("don't know how to handle me class: %s", msgObj.EntityClass.String())
		return nil, nil
	}

}

func CreateOnu2gResponse(tid uint16) ([]byte, error) {
	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(tid), 16),
	}).Info("CreateOnu2gResponse")
	response := &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.Onu2GClassID,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": 0,
			"EquipmentId":     toOctets("BBSM", 20),
			"OpticalNetworkUnitManagementAndControlChannelOmccVersion": 180,
			"VendorProductCode":                           0,
			"SecurityCapability":                          1,
			"SecurityMode":                                1,
			"TotalPriorityQueueNumber":                    1,
			"TotalTrafficSchedulerNumber":                 1,
			"Deprecated":                                  1,
			"TotalGemPortIdNumber":                        32,
			"Sysuptime":                                   319389947, // NOTE need to be smarter?
			"ConnectivityCapability":                      127,
			"CurrentConnectivityMode":                     5,
			"QualityOfServiceQosConfigurationFlexibility": 48,
			"PriorityQueueScaleFactor":                    1,
		},
		Result:        1,
		AttributeMask: 0xfffc,
	}
	pkt, err := serialize(omci.GetResponseType, response, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot serialize Onu2gResponse")
		return nil, err
	}
	return pkt, nil
}

func toOctets(str string, size int) []byte {
	asciiBytes := []byte(str)

	if len(asciiBytes) < size {
		missing := size - len(asciiBytes)
		for i := 0; i < missing; i++ {
			asciiBytes = append(asciiBytes, []byte{0x00}[0])
		}
	}
	return asciiBytes
}
