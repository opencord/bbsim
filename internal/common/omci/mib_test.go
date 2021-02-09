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

// used to verify that the first message in the MIB Sync (OnuData)
// reports the correct MDS
const MDS = uint8(128)

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
	entityID      uint16
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

			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 1, entityID: 0, entityClass: me.OnuDataClassID, attributes: map[string]interface{}{"MibDataSync": MDS}}},
		{"mibUploadNext-1", createTestMibUploadNextArgs(t, 2, 1),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 2, entityID: 257, entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"Type": uint8(47), "NumberOfPorts": uint8(4)}}},
		{"mibUploadNext-4", createTestMibUploadNextArgs(t, 3, 4),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 3, entityID: 257, entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"PowerShedOverride": uint32(0)}}},
		{"mibUploadNext-10", createTestMibUploadNextArgs(t, 4, 10),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 4, entityID: 258, entityClass: me.PhysicalPathTerminationPointEthernetUniClassID, attributes: map[string]interface{}{"SensedType": uint8(47)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create the packet starting from the mibUploadNextRequest
			data, _ := CreateMibUploadNextResponse(tt.args.omciPkt, tt.args.omciMsg, MDS)
			omciMsg, omciPkt := omciBytesToMsg(t, data)

			assert.Equal(t, omciMsg.MessageType, tt.want.messageType)

			msgLayer := (*omciPkt).Layer(omci.LayerTypeMibUploadNextResponse)
			msgObj, msgOk := msgLayer.(*omci.MibUploadNextResponse)
			if !msgOk {
				t.Fail()
			}

			assert.Equal(t, omciMsg.TransactionID, tt.want.transactionId) // tid

			assert.Equal(t, msgObj.ReportedME.GetEntityID(), tt.want.entityID)

			for k, v := range tt.want.attributes {
				attr, _ := msgObj.ReportedME.GetAttribute(k)
				assert.Equal(t, attr, v)
			}
		})
	}
}

type pqueueExpected struct {
	entityId    uint16
	relatedPort uint32
}

func TestGeneratePriorityQueueMe(t *testing.T) {

	tests := []struct {
		name     string
		sequence uint16
		want     pqueueExpected
	}{
		{"generate-pq-downstream-1", 26,
			pqueueExpected{entityId: 1, relatedPort: 16842752}},
		{"generate-pq-downstream-2", 30,
			pqueueExpected{entityId: 2, relatedPort: 16842753}},
		{"generate-pq-downstream-3", 58,
			pqueueExpected{entityId: 9, relatedPort: 16842760}},
		{"generate-pq-upstream-1", 28,
			pqueueExpected{entityId: 32769, relatedPort: 2147549184}},
		{"generate-pq-upstream-2", 32,
			pqueueExpected{entityId: 32770, relatedPort: 2147549185}},
		{"generate-pq-upstream-3", 60,
			pqueueExpected{entityId: 32777, relatedPort: 2147614720}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportedMe, meErr := GeneratePriorityQueueMe(tt.sequence)
			if meErr.GetError() != nil {
				t.Fatal(meErr.Error())
			}

			assert.Equal(t, reportedMe.GetEntityID(), tt.want.entityId)

			relatedPort, _ := reportedMe.GetAttribute("RelatedPort")
			assert.Equal(t, relatedPort, tt.want.relatedPort)
		})
	}

	// test that the related ports are unique
	allRelatedPorts := make(map[uint32]struct{})
	for v := 26; v <= 281; v++ {
		reportedMe, meErr := GeneratePriorityQueueMe(uint16(v))
		if meErr.GetError() != nil {
			t.Fatal(meErr.Error())
		}
		relatedPort, _ := reportedMe.GetAttribute("RelatedPort")
		allRelatedPorts[relatedPort.(uint32)] = struct{}{}
	}

	// we report 128 queues total, but each of them is comprised of 2 messages
	// that's why the 256 iterations
	assert.Equal(t, len(allRelatedPorts), 128)
}
