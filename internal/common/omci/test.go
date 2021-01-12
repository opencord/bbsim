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
)

func ParseTestRequest(omciPkt gopacket.Packet) (*omci.TestRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeGetRequest)
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

// Return true if msg is an Omci Test Request
func IsTestRequest(payload []byte) (bool, error) {
	_, omciMsg, err := ParseOpenOltOmciPacket(payload)
	if err != nil {
		return false, err
	}

	return omciMsg.MessageType == omci.TestRequestType, nil
}

func BuildTestResult(payload []byte) ([]byte, error) {

	omciPkt, omciMsg, err := ParseOpenOltOmciPacket(payload)

	//transactionId, deviceId, _, class, instance, _, err := omcisim.ParsePkt(payload)

	if err != nil {
		return []byte{}, err
	}

	testRequest, err := ParseTestRequest(omciPkt)
	if err != nil {
		return []byte{}, err
	}

	// TODO create a TestResponse using omci-lib-go
	resp := make([]byte, 48)
	resp[0] = byte(omciMsg.TransactionID >> 8)
	resp[1] = byte(omciMsg.TransactionID & 0xFF)
	resp[2] = 27 // Upper nibble 0x0 is fixed (0000), Lower nibbles defines msg type (TestResult=27)
	resp[3] = byte(omciMsg.DeviceIdentifier)
	resp[4] = byte(omciMsg.MessageType)
	resp[5] = byte(omciMsg.MessageType & 0xFF)
	resp[6] = byte(testRequest.EntityInstance >> 8)
	resp[7] = byte(testRequest.EntityInstance & 0xFF)
	// Each of these is a 1-byte code
	// follow by a 2-byte (high, low) value
	resp[8] = 1 // power feed voltage
	resp[9] = 0
	resp[10] = 123 // 123 mV, 20 mv res --> 6mv
	resp[11] = 3   // received optical power
	resp[12] = 1
	resp[13] = 200 // 456 decibel-microwatts, 0.002 dB res --> 0.912 db-mw
	resp[14] = 5   // mean optical launch power
	resp[15] = 3
	resp[16] = 21 // 789 uA, 0.002 dB res --> 1.578 db-mw
	resp[17] = 9  // laser bias current
	resp[18] = 3
	resp[19] = 244 // 1012 uA, 2uA res --> 505 ua
	resp[20] = 12  // temperature
	resp[21] = 38
	resp[22] = 148 // 9876 deg C, 1/256 resolution --> 38.57 Deg C

	return resp, nil
}
