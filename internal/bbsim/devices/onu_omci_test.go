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

package devices

import (
	"github.com/google/gopacket"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	"gotest.tools/assert"
	"testing"
)

var mockAttr = me.AttributeValueMap{
	"PortId":                              0,
	"TContPointer":                        0,
	"Direction":                           0,
	"TrafficManagementPointerForUpstream": 0,
	"TrafficDescriptorProfilePointerForUpstream":   0,
	"PriorityQueuePointerForDownStream":            0,
	"TrafficDescriptorProfilePointerForDownstream": 0,
	"EncryptionKeyRing":                            0,
}

func makeOmciCreateRequest(t *testing.T) []byte {
	omciReq := &omci.CreateRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
		Attributes: mockAttr,
	}

	omciPkt, err := omcilib.Serialize(omci.CreateRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciSetRequest(t *testing.T) []byte {
	omciReq := &omci.SetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
		Attributes: mockAttr,
	}
	omciPkt, err := omcilib.Serialize(omci.SetRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciDeleteRequest(t *testing.T) []byte {
	omciReq := &omci.DeleteRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.DeleteRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciMibResetRequest(t *testing.T) []byte {
	omciReq := &omci.MibResetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.MibResetRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciMessage(t *testing.T, onu *Onu, pkt []byte) bbsim.OmciMessage {
	return bbsim.OmciMessage{
		OnuSN: onu.SerialNumber,
		OnuID: onu.ID,
		OmciMsg: &openolt.OmciMsg{
			IntfId: onu.PonPortID,
			OnuId:  onu.ID,
			Pkt:    pkt,
		},
	}
}

func omciBytesToMsg(t *testing.T, data []byte) (*omci.OMCI, *gopacket.Packet) {
	packet := gopacket.NewPacket(data, omci.LayerTypeOMCI, gopacket.NoCopy)
	if packet == nil {
		t.Fatal("could not decode rxMsg as OMCI")
	}
	omciLayer := packet.Layer(omci.LayerTypeOMCI)
	if omciLayer == nil {
		t.Fatal("could not decode omci layer")
	}
	omciMsg, ok := omciLayer.(*omci.OMCI)
	if !ok {
		t.Fatal("could not assign omci layer")
	}
	return omciMsg, &packet
}

func omciToCreateResponse(t *testing.T, omciPkt *gopacket.Packet) *omci.CreateResponse {
	msgLayer := (*omciPkt).Layer(omci.LayerTypeCreateResponse)
	if msgLayer == nil {
		t.Fatal("omci Msg layer could not be detected for CreateResponse - handling of MibSyncChan stopped")
	}
	msgObj, msgOk := msgLayer.(*omci.CreateResponse)
	if !msgOk {
		t.Fatal("omci Msg layer could not be assigned for CreateResponse - handling of MibSyncChan stopped")
	}
	return msgObj
}

func Test_MibDataSyncIncrease(t *testing.T) {
	onu := createMockOnu(1, 1)

	assert.Equal(t, onu.MibDataSync, uint8(0))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// send a Create and check that MDS has been increased
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// send a Set and check that MDS has been increased
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciSetRequest(t)), stream)
	assert.Equal(t, onu.MibDataSync, uint8(2))

	// send a Delete and check that MDS has been increased
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciDeleteRequest(t)), stream)
	assert.Equal(t, onu.MibDataSync, uint8(3))

	// TODO once supported MDS should increase for:
	// - Start software download
	// - End software download
	// - Activate software
	// - Commit software
}

func Test_MibDataSyncReset(t *testing.T) {
	onu := createMockOnu(1, 1)
	onu.MibDataSync = 192
	assert.Equal(t, onu.MibDataSync, uint8(192))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// send a MibReset and check that MDS has reset to 0
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciMibResetRequest(t)), stream)
	assert.Equal(t, onu.MibDataSync, uint8(0))
}

func Test_MibDataSyncRotation(t *testing.T) {
	onu := createMockOnu(1, 1)
	onu.MibDataSync = 255
	assert.Equal(t, onu.MibDataSync, uint8(255))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// send a request that increases the MDS, but once we're at 255 we should go back to 0 (8bit)
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciDeleteRequest(t)), stream)
	assert.Equal(t, onu.MibDataSync, uint8(0))
}

func Test_GemPortValidation(t *testing.T) {

	// setup
	onu := createMockOnu(1, 1)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// create a gem port via OMCI (gemPortId 12)
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)

	// the first time we created the gemPort
	// the MDS should be incremented
	assert.Equal(t, stream.CallCount, 1)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// and the OMCI response status should be me.Success
	indication := stream.Calls[1].GetOmciInd()
	_, omciPkt := omciBytesToMsg(t, indication.Pkt)
	responseLayer := omciToCreateResponse(t, omciPkt)
	assert.Equal(t, responseLayer.Result, me.Success)

	// send a request to create the same gem port via OMCI (gemPortId 12)
	onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)

	// this time the MDS should not be incremented
	assert.Equal(t, stream.CallCount, 2)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// and the OMCI response status should be me.ProcessingError
	_, omciPkt = omciBytesToMsg(t, stream.Calls[2].GetOmciInd().Pkt)
	responseLayer = omciToCreateResponse(t, omciPkt)
	assert.Equal(t, responseLayer.Result, me.ProcessingError)
}
