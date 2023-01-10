/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	"gotest.tools/assert"
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

func TestSetTxIdInEncodedPacket(t *testing.T) {
	basePkt := []byte{129, 42, 46, 10, 0, 2, 0, 0, 0, 37, 0, 1, 128, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 0, 0, 0, 40, 40, 206, 0, 226}
	res := SetTxIdInEncodedPacket(basePkt, uint16(292))
	assert.Equal(t, len(basePkt), len(res))

	b := res[0:2]
	txId := binary.BigEndian.Uint16(b)
	assert.Equal(t, uint16(292), txId)
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
		NumberOfBaselineCommands: 4,
		baselineItems:            []MibDbEntry{},
	}

	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.OnuDataClassID,
		onuDataEntityId,
		me.AttributeValueMap{me.OnuData_MibDataSync: 0},
		nil,
	})

	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_Type:          ethernetUnitType,
			me.CircuitPack_NumberOfPorts: uniPortCount,
			me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
		},
		nil,
	})

	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.AniGClassID,
		anigEntityId,
		me.AttributeValueMap{
			me.AniG_Arc:                         0,
			me.AniG_ArcInterval:                 0,
			me.AniG_Deprecated:                  0,
			me.AniG_GemBlockLength:              48,
			me.AniG_LowerOpticalThreshold:       255,
			me.AniG_LowerTransmitPowerThreshold: 129,
			me.AniG_OnuResponseTime:             0,
			me.AniG_OpticalSignalLevel:          57428,
			me.AniG_PiggybackDbaReporting:       0,
			me.AniG_SignalDegradeThreshold:      9,
			me.AniG_SignalFailThreshold:         5,
			me.AniG_SrIndication:                1,
			me.AniG_TotalTcontNumber:            8,
			me.AniG_TransmitOpticalLevel:        3171,
			me.AniG_UpperOpticalThreshold:       255,
			me.AniG_UpperTransmitPowerThreshold: 129,
		},
		nil,
	})

	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.Onu2GClassID,
		onu2gEntityId,
		me.AttributeValueMap{
			me.Onu2G_ConnectivityCapability:                      127,
			me.Onu2G_CurrentConnectivityMode:                     0,
			me.Onu2G_Deprecated:                                  1,
			me.Onu2G_PriorityQueueScaleFactor:                    1,
			me.Onu2G_QualityOfServiceQosConfigurationFlexibility: 63,
			me.Onu2G_Sysuptime:                                   0,
			me.Onu2G_TotalGemPortIdNumber:                        8,
			me.Onu2G_TotalPriorityQueueNumber:                    64,
			me.Onu2G_TotalTrafficSchedulerNumber:                 8,
		},
		nil,
	})

	// create an entry with UnkownAttributes set
	var customPktTxId uint16 = 5
	customPkt := []byte{0, byte(customPktTxId), 46, 10, 0, 2, 0, 0, 0, 37, 0, 1, 128, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 0, 0, 0, 40, 40, 206, 0, 226}
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.ClassID(37),
		nil,
		me.AttributeValueMap{},
		customPkt,
	})

	tests := []struct {
		name string
		args mibArgs
		want mibExpected
	}{
		{"mibUploadNext-0", createTestMibUploadNextArgs(t, 1, 0),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 1, entityID: onuDataEntityId.ToUint16(), entityClass: me.OnuDataClassID, attributes: map[string]interface{}{me.OnuData_MibDataSync: uint8(0)}}},
		{"mibUploadNext-1", createTestMibUploadNextArgs(t, 2, 1),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 2, entityID: circuitPackEntityID.ToUint16(), entityClass: me.CircuitPackClassID, attributes: map[string]interface{}{me.CircuitPack_Type: uint8(47), me.CircuitPack_NumberOfPorts: uint8(4)}}},
		{"mibUploadNext-2", createTestMibUploadNextArgs(t, 3, 2),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 3, entityID: anigEntityId.ToUint16(), entityClass: me.AniGClassID, attributes: map[string]interface{}{me.AniG_GemBlockLength: uint16(48)}}},
		{"mibUploadNext-3", createTestMibUploadNextArgs(t, 4, 3),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: 4, entityID: onuDataEntityId.ToUint16(), entityClass: me.Onu2GClassID, attributes: map[string]interface{}{me.Onu2G_TotalPriorityQueueNumber: uint16(64)}}},
		{"mibUploadNext-unknown-attrs", createTestMibUploadNextArgs(t, 5, 4),
			mibExpected{messageType: omci.MibUploadNextResponseType, transactionId: customPktTxId, entityID: 1, entityClass: me.ClassID(37), attributes: map[string]interface{}{"UnknownAttr_1": []uint8{1}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create the packet starting from the mibUploadNextRequest
			data, err := CreateMibUploadNextResponse(tt.args.omciPkt, tt.args.omciMsg, &mibDb)
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
				if isSlice(attr) {
					expected := v.([]uint8)
					data := attr.([]uint8)
					for i, val := range data {
						assert.Equal(t, expected[i], val)
					}
				} else {
					assert.Equal(t, attr, v)
				}
			}
		})
	}

	// now try to get a non existing command from the DB anche expect an error
	args := createTestMibUploadNextArgs(t, 1, 20)
	_, err := CreateMibUploadNextResponse(args.omciPkt, args.omciMsg, &mibDb)
	assert.Error(t, err, "mibdb-does-not-contain-item")
}

func isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
