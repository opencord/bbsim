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

type createArgs struct {
	omciPkt []byte
	result  me.Results
}

type createWant struct {
	result me.Results
}

func TestCreateResponse(t *testing.T) {

	// generate a CreateRequest packet to create a GemPort
	omciReq := &omci.CreateRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
		Attributes: me.AttributeValueMap{
			"PortId":                              0,
			"TContPointer":                        0,
			"Direction":                           0,
			"TrafficManagementPointerForUpstream": 0,
			"TrafficDescriptorProfilePointerForUpstream":   0,
			"PriorityQueuePointerForDownStream":            0,
			"TrafficDescriptorProfilePointerForDownstream": 0,
			"EncryptionKeyRing":                            0,
		},
	}
	omciPkt, err := Serialize(omci.CreateRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = HexEncode(omciPkt)

	tests := []struct {
		name string
		args createArgs
		want createWant
	}{
		{"createSuccess",
			createArgs{omciPkt, me.Success},
			createWant{me.Success},
		},
		{"createProcessingError",
			createArgs{omciPkt, me.ProcessingError},
			createWant{me.ProcessingError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt, msg, _ := ParseOpenOltOmciPacket(tt.args.omciPkt)
			requestPkt, _ := CreateCreateResponse(pkt, msg, tt.args.result)

			omciMsg, omciPkt := omciBytesToMsg(t, requestPkt)

			assert.Equal(t, omciMsg.MessageType, omci.CreateResponseType)

			getResponseLayer := omciToCreateResponse(t, omciPkt)

			assert.Equal(t, getResponseLayer.Result, tt.want.result)

		})
	}
}
