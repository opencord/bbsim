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
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"gotest.tools/assert"
	"testing"
)

func TestCreateMibResetResponse(t *testing.T) {
	data, _ := CreateMibResetResponse(1)

	omciMsg, omciPkt := omciBytesToMsg(t, data)

	assert.Equal(t, omciMsg.MessageType, omci.MibResetResponseType)

	msgLayer := (*omciPkt).Layer(omci.LayerTypeMibResetResponse)
	msgObj, msgOk := msgLayer.(*omci.MibResetResponse)
	if !msgOk {
		t.Fail()
	}

	assert.Equal(t, msgObj.Result, me.Success)
}

// types for TestCreateMibUploadNextResponse test
type mibArgs struct {
	omciPkt gopacket.Packet
	omciMsg *omci.OMCI
}

type mibExpected struct {
	messageType   omci.MessageType
	transactionId uint16
	entityClass   me.ClassID
	attributes    map[string]interface{}
}

func createTestMibUploadNextArgs(t *testing.T, tid uint16, seqNumber uint16) mibArgs {
	mibUploadNext, _ := CreateMibUploadNextRequest(tid, seqNumber)
	mibUploadNext = hexDecode(mibUploadNext)
	mibUploadNextMsg, mibUploadNextPkt := omciBytesToMsg(t, mibUploadNext)

	return mibArgs{
		omciPkt: *mibUploadNextPkt,
		omciMsg: mibUploadNextMsg,
	}
}

func TestCreateMibUploadNextResponse(t *testing.T) {

	tests := []struct {
		name string
		args mibArgs
		want mibExpected
	}{
		{"mibUploadNext-0", createTestMibUploadNextArgs(t, 1, 0),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 1, entityClass: me.OnuDataClassID, attributes: map[string]interface{}{"MibDataSync": uint8(0)}}},
		{"mibUploadNext-1", createTestMibUploadNextArgs(t, 2, 1),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 2, entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"Type": uint8(47), "NumberOfPorts": uint8(4)}}},
		{"mibUploadNext-4", createTestMibUploadNextArgs(t, 3, 4),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 3, entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"PowerShedOverride": uint32(0)}}},
		{"mibUploadNext-10", createTestMibUploadNextArgs(t, 4, 10),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 4, entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"SensedType": uint8(47)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create the packet starting from the mibUploadNextRequest
			data, _ := CreateMibUploadNextResponse(tt.args.omciPkt, tt.args.omciMsg)
			omciMsg, omciPkt := omciBytesToMsg(t, data)

			assert.Equal(t, omciMsg.MessageType, tt.want.messageType)

			msgLayer := (*omciPkt).Layer(omci.LayerTypeMibUploadNextResponse)
			msgObj, msgOk := msgLayer.(*omci.MibUploadNextResponse)
			if !msgOk {
				t.Fail()
			}

			assert.Equal(t, omciMsg.TransactionID, tt.want.transactionId) // tid
			// GetAttribute("ManagedEntityId") returns nil,
			// msgObj.EntityClass is always OnuDataClassID
			// how do we check this?
			//meId, _ := msgObj.ReportedME.GetAttribute("ManagedEntityId")
			//assert.Equal(t, meId, tt.want.entityClass)
			//assert.Equal(t, msgObj.EntityClass, tt.want.entityClass)

			fmt.Println(msgObj.EntityInstance, msgObj.ReportedME.GetEntityID())

			for k, v := range tt.want.attributes {
				attr, _ := msgObj.ReportedME.GetAttribute(k)
				assert.Equal(t, attr, v)
			}
		})
	}

}
