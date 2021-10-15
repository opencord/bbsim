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
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	"gotest.tools/assert"
	"testing"
)

func omciToStartSoftwareDownloadResponse(t *testing.T, omciPkt *gopacket.Packet) *omci.StartSoftwareDownloadResponse {
	msgLayer := (*omciPkt).Layer(omci.LayerTypeStartSoftwareDownloadResponse)
	if msgLayer == nil {
		t.Fatal("omci Msg layer could not be detected for StartSoftwareDownloadResponse")
	}
	msgObj, msgOk := msgLayer.(*omci.StartSoftwareDownloadResponse)
	if !msgOk {
		t.Fatal("omci Msg layer could not be assigned for StartSoftwareDownloadResponse")
	}
	return msgObj
}

func TestCreateStartSoftwareDownloadResponse(t *testing.T) {
	omciReq := &omci.StartSoftwareDownloadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.SoftwareImageClassID,
			EntityInstance: 1,
		},
		ImageSize:            32768,
		NumberOfCircuitPacks: 1,
		WindowSize:           31,
		CircuitPacks:         []uint16{0},
	}

	omciReqPkt, err := Serialize(omci.StartSoftwareDownloadRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciReqPkt, _ = HexEncode(omciReqPkt)

	// start test
	pkt, msg, _ := ParseOpenOltOmciPacket(omciReqPkt)
	responsePkt, err := CreateStartSoftwareDownloadResponse(pkt, msg)
	assert.NilError(t, err)

	omciResponseMsg, omciResponsePkt := omciBytesToMsg(t, responsePkt)
	assert.Equal(t, omciResponseMsg.MessageType, omci.StartSoftwareDownloadResponseType)

	getResponseLayer := omciToStartSoftwareDownloadResponse(t, omciResponsePkt)

	assert.Equal(t, getResponseLayer.Result, me.Success)
}

func TestComputeDownloadSectionsCount(t *testing.T) {
	omciReq := &omci.StartSoftwareDownloadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.SoftwareImageClassID,
			EntityInstance: 1,
		},
		ImageSize:            32768,
		NumberOfCircuitPacks: 1,
		WindowSize:           31,
		CircuitPacks:         []uint16{0},
	}

	omciReqPkt, err := Serialize(omci.StartSoftwareDownloadRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciReqPkt, _ = HexEncode(omciReqPkt)
	pkt, _, _ := ParseOpenOltOmciPacket(omciReqPkt)

	count := ComputeDownloadSectionsCount(pkt)
	assert.Equal(t, count, 1058)
}
