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
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

func ParseTestRequest(omciPkt gopacket.Packet) (*omci.TestRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeTestRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeTestRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.TestRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeTestRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateTestResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error, me.ClassID, uint16, me.Results) {

	msgObj, err := ParseTestRequest(omciPkt)

	if err != nil {
		return nil, err, 0, 0, me.ParameterError
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
	}).Trace("received-test-request")

	var response *omci.TestResponse
	var omciResult me.Results

	switch msgObj.EntityClass {
	case me.AniGClassID:
		omciResult = me.Success
		response = &omci.TestResponse{
			MeBasePacket: omci.MeBasePacket{
				EntityClass:    msgObj.EntityClass,
				EntityInstance: msgObj.EntityInstance,
			},
			Result: omciResult,
		}
	default:
		omciLogger.WithFields(log.Fields{
			"EntityClass": msgObj.EntityClass,
		}).Error("unsupported-me-for-self-test-request")
		omciResult = me.Unsupported
		response = &omci.TestResponse{
			MeBasePacket: omci.MeBasePacket{
				EntityClass:    msgObj.EntityClass,
				EntityInstance: msgObj.EntityInstance,
			},
			Result: omciResult,
		}
	}

	pkt, err := Serialize(omci.TestResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-Serialize-test-response")
		return nil, err, 0, 0, me.ParameterError
	}

	return pkt, nil, msgObj.EntityClass, msgObj.EntityInstance, omciResult

}

func CreateTestResult(classID me.ClassID, instID uint16, tid uint16) ([]byte, error) {
	var result gopacket.SerializableLayer
	switch classID {
	case me.AniGClassID:
		result = &omci.OpticalLineSupervisionTestResult{
			MeBasePacket: omci.MeBasePacket{
				EntityClass:    classID,
				EntityInstance: instID,
			},
			PowerFeedVoltageType:     uint8(1),
			PowerFeedVoltage:         uint16(163),
			ReceivedOpticalPowerType: uint8(3),
			ReceivedOpticalPower:     uint16(rand.Intn(65000)),
			MeanOpticalLaunchType:    uint8(5),
			MeanOpticalLaunch:        uint16(rand.Intn(3000)),
			LaserBiasCurrentType:     uint8(9),
			LaserBiasCurrent:         uint16(rand.Intn(10000)),
			TemperatureType:          uint8(12),
			Temperature:              uint16(rand.Intn(20000)),
			GeneralPurposeBuffer:     uint16(0),
		}
	default:
		return nil, fmt.Errorf("unsupported-class-id-for-test-result--class-id-%v", classID)
	}

	pkt, err := Serialize(omci.TestResultType, result, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-Serialize-test-result")
		return nil, err
	}
	return pkt, nil
}
