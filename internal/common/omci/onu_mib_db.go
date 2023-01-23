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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/gopacket"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
)

// MibDbEntry contains all the information needed to build a
// MibUploadNextResponse packet.
// if Packet has a value all the other fields are ignored and the packet is sent as is.
type MibDbEntry struct {
	classId  me.ClassID
	entityId EntityID
	params   me.AttributeValueMap
	packet   []byte
}

type MibDb struct {
	NumberOfBaselineCommands uint16
	NumberOfExtendedCommands uint16
	baselineItems            []MibDbEntry
	extendedResponses        [][]byte
}

type EntityID []byte

const (
	unknownMePktReportedMeHdr     string = "002500018000"
	unknownMePktAttributes        string = "0102030405060708090A0B0C0D0E0F101112131415161718191A"
	unknownAttribPktReportedMeHdr string = "0101000007FD"
	unknownAttribPktAttributes    string = "00400801000800000006007F07003F0001000100010001000000"
	extRespMsgContentsLenStart           = 8
	extRespMsgContentsLenEnd             = 10
)

func (e EntityID) ToString() string {
	return hex.EncodeToString(e)
}

func (e EntityID) ToUint16() uint16 {
	return binary.BigEndian.Uint16(e)
}

func (e EntityID) ToUint32() uint32 {
	return binary.BigEndian.Uint32(e)
}

func (e EntityID) FromUint16(id uint16) EntityID {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, id)
	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}

func (e EntityID) Equals(i EntityID) bool {
	if res := bytes.Compare(e, i); res == 0 {
		return true
	}
	return false
}

const (
	cardHolderOnuType byte = 0x01 // ONU is a single piece of integrated equipment
	ethernetUnitType  byte = 0x2f // Ethernet BASE-T
	xgsPonUnitType    byte = 0xee // XG-PON10G10
	gPonUnitType      byte = 0xf5 // GPON12441244
	potsUnitType      byte = 0x20 // POTS
	cardHolderSlotID  byte = 0x01
	tcontSlotId       byte = 0x80 // why is this not the same as the cardHolderSlotID, it does not point to anything
	aniGId            byte = 0x01

	upstreamPriorityQueues   = 8 // Number of queues for each T-CONT
	downstreamPriorityQueues = 8 // Number of queues for each PPTP
	tconts                   = 8 // NOTE will we ever need to configure this?
	// trafficSchedulers        = 8  // NOTE will we ever need to configure this?
)

var (
	cardHolderEntityID  = EntityID{cardHolderOnuType, cardHolderSlotID}
	circuitPackEntityID = cardHolderEntityID // is the same as that of the cardholder ME containing this circuit pack instance
)

func GenerateUniPortEntityId(id uint32) EntityID {
	return EntityID{cardHolderSlotID, byte(id)}
}

// creates a MIB database for a ONU
// CircuitPack and CardHolder are static, everything else can be configured
func GenerateMibDatabase(ethUniPortCount int, potsUniPortCount int, technology common.PonTechnology) (*MibDb, error) {

	mibDb := MibDb{
		baselineItems: []MibDbEntry{},
	}

	// the first element to return is the ONU-Data
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.OnuDataClassID,
		EntityID{0x00, 0x00},
		me.AttributeValueMap{me.OnuData_MibDataSync: 0}, // FIXME this needs to be parametrized before sending the response
		nil,
	})

	// then we report the CardHolder
	// NOTE we have not report it till now, so leave it commented out
	//mibDb.items = append(mibDb.items, MibDbEntry{
	//	me.CardholderClassID,
	//	cardHolderEntityID,
	//	me.AttributeValueMap{
	//		"ActualPlugInUnitType":   cardHolderOnuType,
	//		"ExpectedPlugInUnitType": ethernetUnitType,
	//	},
	//})

	// ANI circuitPack
	var aniCPType byte

	switch technology {
	case common.XGSPON:
		aniCPType = xgsPonUnitType
	case common.GPON:
		aniCPType = gPonUnitType
	}

	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_Type:          aniCPType,
			me.CircuitPack_NumberOfPorts: 1, // NOTE is this the ANI port? must be
			me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack-ani", 20),
			me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_VendorId:            ToOctets("ONF", 4),
			me.CircuitPack_AdministrativeState: 0,
			me.CircuitPack_OperationalState:    0,
			me.CircuitPack_BridgedOrIpInd:      0,
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_CardConfiguration:           0,
			me.CircuitPack_TotalTContBufferNumber:      8,
			me.CircuitPack_TotalPriorityQueueNumber:    8,
			me.CircuitPack_TotalTrafficSchedulerNumber: 0,
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_PowerShedOverride: uint32(0),
		},
		nil,
	})

	// ANI-G
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.AniGClassID,
		EntityID{tcontSlotId, aniGId},
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

	// circuitPack Ethernet
	// NOTE the circuit pack is divided in multiple messages as too big to fit in a single one
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_Type:          ethernetUnitType,
			me.CircuitPack_NumberOfPorts: ethUniPortCount,
			me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_VendorId:            ToOctets("ONF", 4),
			me.CircuitPack_AdministrativeState: 0,
			me.CircuitPack_OperationalState:    0,
			me.CircuitPack_BridgedOrIpInd:      0,
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_CardConfiguration:           0,
			me.CircuitPack_TotalTContBufferNumber:      8,
			me.CircuitPack_TotalPriorityQueueNumber:    8,
			me.CircuitPack_TotalTrafficSchedulerNumber: 16,
		},
		nil,
	})
	mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_PowerShedOverride: uint32(0),
		},
		nil,
	})

	if potsUniPortCount > 0 {
		// circuitPack POTS
		// NOTE the circuit pack is divided in multiple messages as too big to fit in a single one
		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_Type:          potsUnitType,
				me.CircuitPack_NumberOfPorts: potsUniPortCount,
				me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack", 20),
				me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
			},
			nil,
		})
		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_VendorId:            ToOctets("ONF", 4),
				me.CircuitPack_AdministrativeState: 0,
				me.CircuitPack_OperationalState:    0,
				me.CircuitPack_BridgedOrIpInd:      0,
			},
			nil,
		})
		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
				me.CircuitPack_CardConfiguration:           0,
				me.CircuitPack_TotalTContBufferNumber:      8,
				me.CircuitPack_TotalPriorityQueueNumber:    8,
				me.CircuitPack_TotalTrafficSchedulerNumber: 16,
			},
			nil,
		})
		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_PowerShedOverride: uint32(0),
			},
			nil,
		})
	}

	// PPTP and UNI-Gs
	// NOTE this are dependent on the number of UNI this ONU supports
	// Through an identical ID, the UNI-G ME is implicitly linked to an instance of a PPTP
	totalPortsCount := ethUniPortCount + potsUniPortCount
	for i := 1; i <= totalPortsCount; i++ {
		uniEntityId := GenerateUniPortEntityId(uint32(i))

		if i <= ethUniPortCount {
			// first, create the correct amount of ethernet UNIs, the same is done in onu.go
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PhysicalPathTerminationPointEthernetUniClassID,
				uniEntityId,
				me.AttributeValueMap{
					me.PhysicalPathTerminationPointEthernetUni_ExpectedType:                  0,
					me.PhysicalPathTerminationPointEthernetUni_SensedType:                    ethernetUnitType,
					me.PhysicalPathTerminationPointEthernetUni_AutoDetectionConfiguration:    0,
					me.PhysicalPathTerminationPointEthernetUni_EthernetLoopbackConfiguration: 0,
					me.PhysicalPathTerminationPointEthernetUni_AdministrativeState:           0,
					me.PhysicalPathTerminationPointEthernetUni_OperationalState:              0,
					me.PhysicalPathTerminationPointEthernetUni_ConfigurationInd:              3,
					me.PhysicalPathTerminationPointEthernetUni_MaxFrameSize:                  1518,
					me.PhysicalPathTerminationPointEthernetUni_DteOrDceInd:                   0,
					me.PhysicalPathTerminationPointEthernetUni_PauseTime:                     0,
					me.PhysicalPathTerminationPointEthernetUni_BridgedOrIpInd:                2,
					me.PhysicalPathTerminationPointEthernetUni_Arc:                           0,
					me.PhysicalPathTerminationPointEthernetUni_ArcInterval:                   0,
					me.PhysicalPathTerminationPointEthernetUni_PppoeFilter:                   0,
					me.PhysicalPathTerminationPointEthernetUni_PowerControl:                  0,
				},
				nil,
			})
		} else {
			// the remaining ones are pots UNIs, the same is done in onu.go
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PhysicalPathTerminationPointPotsUniClassID,
				uniEntityId,
				me.AttributeValueMap{
					me.PhysicalPathTerminationPointPotsUni_AdministrativeState: 0,
					me.PhysicalPathTerminationPointPotsUni_Deprecated:          0,
					me.PhysicalPathTerminationPointPotsUni_Arc:                 0,
					me.PhysicalPathTerminationPointPotsUni_ArcInterval:         0,
					me.PhysicalPathTerminationPointPotsUni_Impedance:           0,
					me.PhysicalPathTerminationPointPotsUni_TransmissionPath:    0,
					me.PhysicalPathTerminationPointPotsUni_RxGain:              0,
					me.PhysicalPathTerminationPointPotsUni_TxGain:              0,
					me.PhysicalPathTerminationPointPotsUni_OperationalState:    0,
					me.PhysicalPathTerminationPointPotsUni_HookState:           0,
					me.PhysicalPathTerminationPointPotsUni_PotsHoldoverTime:    0,
					me.PhysicalPathTerminationPointPotsUni_NominalFeedVoltage:  0,
					me.PhysicalPathTerminationPointPotsUni_LossOfSoftswitch:    0,
				},
				nil,
			})
		}

		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.UniGClassID,
			uniEntityId,
			me.AttributeValueMap{
				me.UniG_AdministrativeState:         0,
				me.UniG_Deprecated:                  0,
				me.UniG_ManagementCapability:        0,
				me.UniG_NonOmciManagementIdentifier: 0,
				me.UniG_RelayAgentOptions:           0,
			},
			nil,
		})

		// Downstream Queues (related to PPTP)
		// downstreamPriorityQueues for each UNI Port
		// EntityID = MSB: cardHolderSlotID, LSB: uniPortNo<<4 + prio
		for j := 0; j < downstreamPriorityQueues; j++ {
			queueEntityId := EntityID{cardHolderSlotID, byte(i<<4 + j)}

			// we first report the PriorityQueue without any attribute
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{},
				nil,
			})

			// then we report it with the required attributes
			// In the downstream direction, the first byte is the slot number and the second byte is the port number of the queue's destination port.
			relatedPort := append(uniEntityId, 0x00, byte(j))
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{
					me.PriorityQueue_QueueConfigurationOption:                            0,
					me.PriorityQueue_MaximumQueueSize:                                    100,
					me.PriorityQueue_AllocatedQueueSize:                                  100,
					me.PriorityQueue_DiscardBlockCounterResetInterval:                    0,
					me.PriorityQueue_ThresholdValueForDiscardedBlocksDueToBufferOverflow: 0,
					me.PriorityQueue_RelatedPort:                                         relatedPort.ToUint32(),
					me.PriorityQueue_TrafficSchedulerPointer:                             0, //it was hardcoded to 0x0108 in the current implementation
					me.PriorityQueue_Weight:                                              1,
					me.PriorityQueue_BackPressureOperation:                               1,
					me.PriorityQueue_BackPressureTime:                                    0,
					me.PriorityQueue_BackPressureOccurQueueThreshold:                     0,
					me.PriorityQueue_BackPressureClearQueueThreshold:                     0,
				},
				nil,
			})
		}
	}

	// T-CONTS and Traffic Schedulers
	for i := 1; i <= tconts; i++ {
		tcontEntityId := EntityID{tcontSlotId, byte(i)}

		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.TContClassID,
			tcontEntityId,
			me.AttributeValueMap{
				me.TCont_AllocId: 65535,
			},
			nil,
		})

		tsEntityId := EntityID{cardHolderSlotID, byte(i)}
		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.TrafficSchedulerClassID,
			tsEntityId, //was not reported in the original implementation
			me.AttributeValueMap{
				me.TrafficScheduler_TContPointer:            tcontEntityId.ToUint16(), // was hardcoded to a non-existing t-cont
				me.TrafficScheduler_TrafficSchedulerPointer: 0,
				me.TrafficScheduler_Policy:                  02,
				me.TrafficScheduler_PriorityWeight:          0,
			},
			nil,
		})

		for j := 0; j < upstreamPriorityQueues; j++ {
			queueEntityId := EntityID{tcontSlotId, byte(i<<4 + j)}
			// Upstream Queues (related to traffic schedulers)
			// upstreamPriorityQueues per TCONT
			// EntityID = MSB: tcontSlotId, LSB: tcontNo<<4 + prio

			// we first report the PriorityQueue without any attribute
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{},
				nil,
			})

			// then we report it with the required attributes
			// In the upstream direction, the first 2 bytes are the ME ID of the associated T- CONT, the first byte of which is a slot number, the second byte a T-CONT number.
			relatedPort := append(tcontEntityId, 0x00, byte(j))
			mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{
					me.PriorityQueue_QueueConfigurationOption:                            0,
					me.PriorityQueue_MaximumQueueSize:                                    100,
					me.PriorityQueue_AllocatedQueueSize:                                  100,
					me.PriorityQueue_DiscardBlockCounterResetInterval:                    0,
					me.PriorityQueue_ThresholdValueForDiscardedBlocksDueToBufferOverflow: 0,
					me.PriorityQueue_RelatedPort:                                         relatedPort.ToUint32(),
					me.PriorityQueue_TrafficSchedulerPointer:                             tsEntityId.ToUint16(), //it was hardcoded to 0x0108 in the current implementation
					me.PriorityQueue_Weight:                                              1,
					me.PriorityQueue_BackPressureOperation:                               1,
					me.PriorityQueue_BackPressureTime:                                    0,
					me.PriorityQueue_BackPressureOccurQueueThreshold:                     0,
					me.PriorityQueue_BackPressureClearQueueThreshold:                     0,
				},
				nil,
			})
		}
	}

	// ONU-2g

	onu2g := MibDbEntry{
		me.Onu2GClassID,
		EntityID{0x00, 0x00},
		me.AttributeValueMap{
			//me.Onu2G_EquipmentId: 1,
			//me.Onu2G_OpticalNetworkUnitManagementAndControlChannelOmccVersion: 2,
			//me.Onu2G_VendorProductCode: 3,
			//me.Onu2G_SecurityCapability: 4,
			//me.Onu2G_SecurityMode: 5,

			me.Onu2G_TotalPriorityQueueNumber:                    64,
			me.Onu2G_TotalTrafficSchedulerNumber:                 8,
			me.Onu2G_Deprecated:                                  1,
			me.Onu2G_TotalGemPortIdNumber:                        8,
			me.Onu2G_Sysuptime:                                   6,
			me.Onu2G_ConnectivityCapability:                      127,
			me.Onu2G_CurrentConnectivityMode:                     7,
			me.Onu2G_QualityOfServiceQosConfigurationFlexibility: 63,
			me.Onu2G_PriorityQueueScaleFactor:                    1,
		},
		nil,
	}

	if common.Config.BBSim.InjectOmciUnknownAttributes {
		// NOTE the TxID is actually replaced
		// by SetTxIdInEncodedPacket in CreateMibUploadNextResponse
		txId := uint16(33066)

		b := make([]byte, 4)
		binary.BigEndian.PutUint16(b, txId)
		b[2] = byte(omci.MibUploadNextResponseType)
		b[3] = byte(omci.BaselineIdent)
		omciHdr := hex.EncodeToString(b)

		//omciHdr := "00032e0a"
		msgHdr := "00020000"
		trailer := "0000002828ce00e2"

		msg := omciHdr + msgHdr + unknownAttribPktReportedMeHdr + unknownAttribPktAttributes + trailer
		data, err := hex.DecodeString(msg)
		if err != nil {
			omciLogger.Fatal("cannot-create-custom-packet")
		}
		packet := gopacket.NewPacket(data, omci.LayerTypeOMCI, gopacket.Lazy)

		onu2g = MibDbEntry{
			me.Onu2GClassID,
			nil,
			me.AttributeValueMap{},
			packet.Data(),
		}
	}

	mibDb.baselineItems = append(mibDb.baselineItems, onu2g)

	if common.Config.BBSim.InjectOmciUnknownMe {
		// NOTE the TxID is actually replaced
		// by SetTxIdInEncodedPacket in CreateMibUploadNextResponse
		txId := uint16(33066)

		b := make([]byte, 4)
		binary.BigEndian.PutUint16(b, txId)
		b[2] = byte(omci.MibUploadNextResponseType)
		b[3] = byte(omci.BaselineIdent)
		omciHdr := hex.EncodeToString(b)

		//omciHdr := "00032e0a"
		msgHdr := "00020000"
		trailer := "0000002828ce00e2"
		msg := omciHdr + msgHdr + unknownMePktReportedMeHdr + unknownMePktAttributes + trailer
		data, err := hex.DecodeString(msg)
		if err != nil {
			omciLogger.Fatal("cannot-create-custom-packet")
		}

		packet := gopacket.NewPacket(data, omci.LayerTypeOMCI, gopacket.Lazy)

		mibDb.baselineItems = append(mibDb.baselineItems, MibDbEntry{
			me.ClassID(37), // G.988 "Intentionally left blank"
			nil,
			me.AttributeValueMap{},
			packet.Data(),
		})
	}

	mibDb.NumberOfBaselineCommands = uint16(len(mibDb.baselineItems))

	// Create extended MIB upload responses
	omciLayer := &omci.OMCI{
		TransactionID:    0xFFFF, //to be replaced later on
		MessageType:      omci.MibUploadNextResponseType,
		DeviceIdentifier: omci.ExtendedIdent,
	}
	var i uint16 = 0
	for i < mibDb.NumberOfBaselineCommands {
		currentEntry := mibDb.baselineItems[i]
		for mibDb.baselineItems[i].packet != nil {
			// Skip any entry with a predefined packet currently used for MEs with unknown ClassID or unknown attributes.
			// This information will be added later to the last extended response packet
			i++
			currentEntry = mibDb.baselineItems[i]
		}
		reportedME, meErr := me.LoadManagedEntityDefinition(currentEntry.classId, me.ParamData{
			EntityID:   currentEntry.entityId.ToUint16(),
			Attributes: currentEntry.params,
		})
		if meErr.GetError() != nil {
			omciLogger.Errorf("Error while generating reportedME %s: %v", currentEntry.classId.String(), meErr.Error())
		}
		request := &omci.MibUploadNextResponse{
			MeBasePacket: omci.MeBasePacket{
				EntityClass:    me.OnuDataClassID,
				EntityInstance: uint16(0),
				Extended:       true,
			},
			ReportedME:    *reportedME,
			AdditionalMEs: make([]me.ManagedEntity, 0),
		}
		i++
		var outgoingPacket []byte
		var outgoingPacketLen = 0
	addMeLoop:
		for outgoingPacketLen <= omci.MaxExtendedLength-omci.MaxBaselineLength && i < mibDb.NumberOfBaselineCommands {
			currentEntry := mibDb.baselineItems[i]
			for mibDb.baselineItems[i].packet != nil {
				// Skip any entry with a predefined packet currently used for MEs with unknown ClassID or unknown attributes.
				// This information will be added later to the last extended response packet
				i++
				if i < mibDb.NumberOfBaselineCommands {
					currentEntry = mibDb.baselineItems[i]
				} else {
					break addMeLoop
				}
			}
			additionalME, meErr := me.LoadManagedEntityDefinition(currentEntry.classId, me.ParamData{
				EntityID:   currentEntry.entityId.ToUint16(),
				Attributes: currentEntry.params,
			})
			if meErr.GetError() != nil {
				omciLogger.Errorf("Error while generating additionalME %s: %v", currentEntry.classId.String(), meErr.Error())
			}
			request.AdditionalMEs = append(request.AdditionalMEs, *additionalME)

			var options gopacket.SerializeOptions
			options.FixLengths = true

			buffer := gopacket.NewSerializeBuffer()
			omciErr := gopacket.SerializeLayers(buffer, options, omciLayer, request)
			if omciErr != nil {
				omciLogger.Errorf("Error while serializing generating additionalME %s: %v", currentEntry.classId.String(), omciErr)
			}
			outgoingPacket = buffer.Bytes()
			outgoingPacketLen = len(outgoingPacket)
			i++
		}
		mibDb.extendedResponses = append(mibDb.extendedResponses, outgoingPacket)
		mibDb.NumberOfExtendedCommands = uint16(len(mibDb.extendedResponses))

		outgoingPacketString := strings.ToLower(hex.EncodeToString(mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1]))
		omciLogger.Debugf("Extended MIB upload response respNo: %d length: %d  string: %s", mibDb.NumberOfExtendedCommands, outgoingPacketLen, outgoingPacketString)
	}
	// Currently, there is enough space in the last extended response to add potential MEs with unknown ClassID or unknown attributes, if requested.
	if common.Config.BBSim.InjectOmciUnknownMe {
		var err error
		mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1], err = AppendAdditionalMEs(mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1], unknownMePktReportedMeHdr, unknownMePktAttributes)
		if err != nil {
			omciLogger.Fatal(err)
		} else {
			outgoingPacketString := strings.ToLower(hex.EncodeToString(mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1]))
			omciLogger.Debugf("Reponse with unknown ME added: %s", outgoingPacketString)
		}
	}
	if common.Config.BBSim.InjectOmciUnknownAttributes {
		var err error
		mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1], err = AppendAdditionalMEs(mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1], unknownAttribPktReportedMeHdr, unknownAttribPktAttributes)
		if err != nil {
			omciLogger.Fatal(err)
		} else {
			outgoingPacketString := strings.ToLower(hex.EncodeToString(mibDb.extendedResponses[mibDb.NumberOfExtendedCommands-1]))
			omciLogger.Debugf("Reponse with unknown attributes added: %s", outgoingPacketString)
		}
	}
	return &mibDb, nil
}

func AppendAdditionalMEs(srcSlice []byte, reportedMeHdr string, attributes string) ([]byte, error) {
	attribBytes, err := hex.DecodeString(attributes)
	if err != nil {
		return nil, fmt.Errorf("cannot-decode-attributes-string")
	}
	attribBytesLen := len(attribBytes)
	attribBytesLenStr := fmt.Sprintf("%04X", attribBytesLen)
	msg := attribBytesLenStr + reportedMeHdr + attributes
	data, err := hex.DecodeString(msg)
	if err != nil {
		return nil, fmt.Errorf("cannot-decode-attributes")
	}
	dstSlice := make([]byte, len(srcSlice))
	copy(dstSlice, srcSlice)
	dstSlice = append(dstSlice[:], data[:]...)
	messageContentsLen := binary.BigEndian.Uint16(dstSlice[extRespMsgContentsLenStart:extRespMsgContentsLenEnd])
	dataLen := len(data)
	newMessageContentsLen := messageContentsLen + uint16(dataLen)
	newLenSlice := make([]byte, 2)
	binary.BigEndian.PutUint16(newLenSlice, newMessageContentsLen)
	copy(dstSlice[extRespMsgContentsLenStart:extRespMsgContentsLenEnd], newLenSlice[0:2])
	return dstSlice, nil
}
