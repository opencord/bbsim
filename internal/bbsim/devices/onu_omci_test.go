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
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	"gotest.tools/assert"
	"testing"
)

var mockAttr = me.AttributeValueMap{
	"ManagedEntityId":                     12,
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
