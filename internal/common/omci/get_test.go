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
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"gotest.tools/assert"
	"testing"
)

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

func omciToGetResponse(t *testing.T, omciPkt *gopacket.Packet) *omci.GetResponse {
	msgLayer := (*omciPkt).Layer(omci.LayerTypeGetResponse)
	if msgLayer == nil {
		t.Fatal("omci Msg layer could not be detected for GetResponse - handling of MibSyncChan stopped")
	}
	msgObj, msgOk := msgLayer.(*omci.GetResponse)
	if !msgOk {
		t.Fatal("omci Msg layer could not be assigned for GetResponse - handling of MibSyncChan stopped")
	}
	return msgObj
}

type args struct {
	generatedPkt  *omci.GetResponse
	transactionId uint16
}

type want struct {
	transactionId uint16
	attributes    map[string]interface{}
}

func TestGetResponse(t *testing.T) {

	tests := []struct {
		name string
		args args
		want want
	}{
		{"getOnu2gResponse",
			args{createOnu2gResponse(57344, 10), 1},
			want{1, map[string]interface{}{"OpticalNetworkUnitManagementAndControlChannelOmccVersion": uint8(180)}},
		},
		{"getOnugResponse",
			args{createOnugResponse(40960, 10), 1},
			want{1, map[string]interface{}{}},
		},
		{"getOnuDataResponse",
			args{createOnuDataResponse(32768, 10, 129), 2},
			want{2, map[string]interface{}{"MibDataSync": uint8(129)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, _ := Serialize(omci.GetResponseType, tt.args.generatedPkt, tt.args.transactionId)

			// emulate the openonu-go behavior:
			// omci_cc.receiveMessage process the message (creates a gopacket and extracts the OMCI layer) and invokes a callback
			// in the GetResponse case omci_cc.receiveOmciResponse
			// then the OmciMessage (gopacket + OMIC layer) is is published on a channel
			omciMsg, omciPkt := omciBytesToMsg(t, data)

			assert.Equal(t, omciMsg.MessageType, omci.GetResponseType)
			assert.Equal(t, omciMsg.TransactionID, tt.want.transactionId)

			// that is read by myb_sync.processMibSyncMessages
			// the myb_sync.handleOmciMessage is called and then
			// myb_sync.handleOmciGetResponseMessage where we extract the GetResponse layer
			getResponseLayer := omciToGetResponse(t, omciPkt)

			assert.Equal(t, getResponseLayer.Result, me.Success)

			for k, v := range tt.want.attributes {
				attr, _ := getResponseLayer.Attributes[k]
				assert.Equal(t, attr, v)
			}
		})
	}
}

func TestCreateOnugResponse(t *testing.T) {
	response := createOnugResponse(40960, 1)
	data, _ := Serialize(omci.GetResponseType, response, 1)

	omciMsg, omciPkt := omciBytesToMsg(t, data)

	assert.Equal(t, omciMsg.MessageType, omci.GetResponseType)

	getResponseLayer := omciToGetResponse(t, omciPkt)

	assert.Equal(t, getResponseLayer.Result, me.Success)
}
