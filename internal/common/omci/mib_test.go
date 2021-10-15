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
	mibUploadNext = HexDecode(mibUploadNext)
	mibUploadNextMsg, mibUploadNextPkt := omciBytesToMsg(t, mibUploadNext)

	return mibArgs{
		omciPkt: *mibUploadNextPkt,
		omciMsg: mibUploadNextMsg,
	}
}

func TestCreateMibUploadNextResponse(t *testing.T) {

	const uniPortCount = 4

	var (
		onuDataEntityId = EntityID{0x00, 0x00}
		onu2gEntityId   = EntityID{0x00, 0x00}
		anigEntityId    = EntityID{tcontSlotId, aniGId}
	)

	// create a fake mibDb, we only need to test that given a CommandSequenceNumber
	// we return the corresponding entry
	// the only exception is for OnuData in which we need to replace the MibDataSync attribute with the current value
	mibDb := MibDb{
		NumberOfCommands: 4,
		items:            []MibDbEntry{},
	}

	mibDb.items = append(mibDb.items, MibDbEntry{
		me.OnuDataClassID,
		onuDataEntityId,
		me.AttributeValueMap{"MibDataSync": 0},
	})

	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"Type":          ethernetUnitType,
			"NumberOfPorts": uniPortCount,
			"SerialNumber":  ToOctets("BBSM-Circuit-Pack", 20),
			"Version":       ToOctets("v0.0.1", 20),
		},
	})

	mibDb.items = append(mibDb.items, MibDbEntry{
		me.AniGClassID,
		anigEntityId,
		me.AttributeValueMap{
			"Arc":                         0,
			"ArcInterval":                 0,
			"Deprecated":                  0,
			"GemBlockLength":              48,
			"LowerOpticalThreshold":       255,
			"LowerTransmitPowerThreshold": 129,
			"OnuResponseTime":             0,
			"OpticalSignalLevel":          57428,
			"PiggybackDbaReporting":       0,
			"SignalDegradeThreshold":      9,
			"SignalFailThreshold":         5,
			"SrIndication":                1,
			"TotalTcontNumber":            8,
			"TransmitOpticalLevel":        3171,
			"UpperOpticalThreshold":       255,
			"UpperTransmitPowerThreshold": 129,
		},
	})

	mibDb.items = append(mibDb.items, MibDbEntry{
		me.Onu2GClassID,
		onu2gEntityId,
		me.AttributeValueMap{
			"ConnectivityCapability":                      127,
			"CurrentConnectivityMode":                     0,
			"Deprecated":                                  1,
			"PriorityQueueScaleFactor":                    1,
			"QualityOfServiceQosConfigurationFlexibility": 63,
			"Sysuptime":                                   0,
			"TotalGemPortIdNumber":                        8,
			"TotalPriorityQueueNumber":                    64,
			"TotalTrafficSchedulerNumber":                 8,
		},
	})

	tests := []struct {
		name string
		args mibArgs
		want mibExpected
	}{
		{"mibUploadNext-0", createTestMibUploadNextArgs(t, 1, 0),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 1, entityID: onuDataEntityId.ToUint16(), entityClass: me.OnuDataClassID, attributes: map[string]interface{}{"MibDataSync": MDS}}},
		{"mibUploadNext-1", createTestMibUploadNextArgs(t, 2, 1),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 2, entityID: circuitPackEntityID.ToUint16(), entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{"Type": uint8(47), "NumberOfPorts": uint8(4)}}},
		{"mibUploadNext-2", createTestMibUploadNextArgs(t, 3, 2),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 3, entityID: anigEntityId.ToUint16(), entityClass: me.AniGClassID, attributes: map[string]interface{}{"GemBlockLength": uint16(48)}}},
		{"mibUploadNext-3", createTestMibUploadNextArgs(t, 4, 3),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 4, entityID: onuDataEntityId.ToUint16(), entityClass: me.Onu2GClassID, attributes: map[string]interface{}{"TotalPriorityQueueNumber": uint16(64)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create the packet starting from the mibUploadNextRequest
			data, err := CreateMibUploadNextResponse(tt.args.omciPkt, tt.args.omciMsg, MDS, &mibDb)
			assert.NilError(t, err)
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
				attr, err := msgObj.ReportedME.GetAttribute(k)
				assert.NilError(t, err)
				assert.Equal(t, attr, v)
			}
		})
	}

	// now try to get a non existing command from the DB anche expect an error
	args := createTestMibUploadNextArgs(t, 1, 20)
	_, err := CreateMibUploadNextResponse(args.omciPkt, args.omciMsg, MDS, &mibDb)
	assert.Error(t, err, "mibdb-does-not-contain-item")
}
