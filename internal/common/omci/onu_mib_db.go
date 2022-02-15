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
	"bytes"
	"encoding/binary"
	"encoding/hex"

	me "github.com/opencord/omci-lib-go/v2/generated"
)

type MibDbEntry struct {
	classId  me.ClassID
	entityId EntityID
	params   me.AttributeValueMap
}

type MibDb struct {
	NumberOfCommands uint16
	items            []MibDbEntry
}

type EntityID []byte

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
	potsUnitType      byte = 0x20 // POTS
	cardHolderSlotID  byte = 0x01
	tcontSlotId       byte = 0x80 // why is this not the same as the cardHolderSlotID, it does not point to anything
	aniGId            byte = 0x01

	upstreamPriorityQueues   = 8  // Number of queues for each T-CONT
	downstreamPriorityQueues = 16 // Number of queues for each PPTP
	tconts                   = 8  // NOTE will we ever need to configure this?
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
func GenerateMibDatabase(ethUniPortCount int, potsUniPortCount int) (*MibDb, error) {

	mibDb := MibDb{
		items: []MibDbEntry{},
	}

	// the first element to return is the ONU-Data
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.OnuDataClassID,
		EntityID{0x00, 0x00},
		me.AttributeValueMap{me.OnuData_MibDataSync: 0}, // FIXME this needs to be parametrized before sending the response
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

	// circuitPack XG-PON10G10
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_Type:          xgsPonUnitType,
			me.CircuitPack_NumberOfPorts: 1, // NOTE is this the ANI port? must be
			me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack-ani", 20),
			me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_VendorId:            ToOctets("ONF", 4),
			me.CircuitPack_AdministrativeState: 0,
			me.CircuitPack_OperationalState:    0,
			me.CircuitPack_BridgedOrIpInd:      0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_CardConfiguration:           0,
			me.CircuitPack_TotalTContBufferNumber:      8,
			me.CircuitPack_TotalPriorityQueueNumber:    8,
			me.CircuitPack_TotalTrafficSchedulerNumber: 0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_PowerShedOverride: uint32(0),
		},
	})

	// ANI-G
	mibDb.items = append(mibDb.items, MibDbEntry{
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
	})

	// circuitPack Ethernet
	// NOTE the circuit pack is divided in multiple messages as too big to fit in a single one
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_Type:          ethernetUnitType,
			me.CircuitPack_NumberOfPorts: ethUniPortCount,
			me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_VendorId:            ToOctets("ONF", 4),
			me.CircuitPack_AdministrativeState: 0,
			me.CircuitPack_OperationalState:    0,
			me.CircuitPack_BridgedOrIpInd:      0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
			me.CircuitPack_CardConfiguration:           0,
			me.CircuitPack_TotalTContBufferNumber:      8,
			me.CircuitPack_TotalPriorityQueueNumber:    8,
			me.CircuitPack_TotalTrafficSchedulerNumber: 16,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			me.CircuitPack_PowerShedOverride: uint32(0),
		},
	})

	if potsUniPortCount > 0 {
		// circuitPack POTS
		// NOTE the circuit pack is divided in multiple messages as too big to fit in a single one
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_Type:          potsUnitType,
				me.CircuitPack_NumberOfPorts: potsUniPortCount,
				me.CircuitPack_SerialNumber:  ToOctets("BBSM-Circuit-Pack", 20),
				me.CircuitPack_Version:       ToOctets("v0.0.1", 20),
			},
		})
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_VendorId:            ToOctets("ONF", 4),
				me.CircuitPack_AdministrativeState: 0,
				me.CircuitPack_OperationalState:    0,
				me.CircuitPack_BridgedOrIpInd:      0,
			},
		})
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_EquipmentId:                 ToOctets("BBSM-Circuit-Pack", 20),
				me.CircuitPack_CardConfiguration:           0,
				me.CircuitPack_TotalTContBufferNumber:      8,
				me.CircuitPack_TotalPriorityQueueNumber:    8,
				me.CircuitPack_TotalTrafficSchedulerNumber: 16,
			},
		})
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.CircuitPackClassID,
			circuitPackEntityID,
			me.AttributeValueMap{
				me.CircuitPack_PowerShedOverride: uint32(0),
			},
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
			mibDb.items = append(mibDb.items, MibDbEntry{
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
			})
		} else {
			// the remaining ones are pots UNIs, the same is done in onu.go
			mibDb.items = append(mibDb.items, MibDbEntry{
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
			})
		}

		mibDb.items = append(mibDb.items, MibDbEntry{
			me.UniGClassID,
			uniEntityId,
			me.AttributeValueMap{
				me.UniG_AdministrativeState:         0,
				me.UniG_Deprecated:                  0,
				me.UniG_ManagementCapability:        0,
				me.UniG_NonOmciManagementIdentifier: 0,
				me.UniG_RelayAgentOptions:           0,
			},
		})

		// Downstream Queues (related to PPTP)
		// 16 priorities queues for each UNI Ports
		// EntityID = cardHolderSlotID + Uni EntityID (0101)
		for j := 1; j <= downstreamPriorityQueues; j++ {
			queueEntityId := EntityID{cardHolderSlotID, byte(j)}

			// we first report the PriorityQueue without any attribute
			mibDb.items = append(mibDb.items, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{},
			})

			// then we report it with the required attributes
			// In the downstream direction, the first byte is the slot number and the second byte is the port number of the queue's destination port.
			relatedPort := append(uniEntityId, 0x00, byte(j))
			mibDb.items = append(mibDb.items, MibDbEntry{
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
			})
		}
	}

	// T-CONTS and Traffic Schedulers
	for i := 1; i <= tconts; i++ {
		tcontEntityId := EntityID{tcontSlotId, byte(i)}

		mibDb.items = append(mibDb.items, MibDbEntry{
			me.TContClassID,
			tcontEntityId,
			me.AttributeValueMap{
				me.TCont_AllocId: 65535,
			},
		})

		tsEntityId := EntityID{cardHolderSlotID, byte(i)}
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.TrafficSchedulerClassID,
			tsEntityId, //was not reported in the original implementation
			me.AttributeValueMap{
				me.TrafficScheduler_TContPointer:            tcontEntityId.ToUint16(), // was hardcoded to a non-existing t-cont
				me.TrafficScheduler_TrafficSchedulerPointer: 0,
				me.TrafficScheduler_Policy:                  02,
				me.TrafficScheduler_PriorityWeight:          0,
			},
		})

		for j := 1; j <= upstreamPriorityQueues; j++ {
			queueEntityId := EntityID{tcontSlotId, byte(j)}
			// Upstream Queues (related to traffic schedulers)
			// 8 priorities queues per TCONT
			// EntityID = tcontSlotId + Uni EntityID (8001)

			// we first report the PriorityQueue without any attribute
			mibDb.items = append(mibDb.items, MibDbEntry{
				me.PriorityQueueClassID,
				queueEntityId, //was not reported in the original implementation
				me.AttributeValueMap{},
			})

			// then we report it with the required attributes
			// In the upstream direction, the first 2 bytes are the ME ID of the associated T- CONT, the first byte of which is a slot number, the second byte a T-CONT number.
			relatedPort := append(tcontEntityId, 0x00, byte(j))
			mibDb.items = append(mibDb.items, MibDbEntry{
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
			})
		}
	}

	// ONU-2g
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.Onu2GClassID,
		EntityID{0x00, 0x00},
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
	})

	mibDb.NumberOfCommands = uint16(len(mibDb.items))

	return &mibDb, nil
}