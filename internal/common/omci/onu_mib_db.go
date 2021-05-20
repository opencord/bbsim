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
	me "github.com/opencord/omci-lib-go/generated"
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
func GenerateMibDatabase(uniPortCount int) (*MibDb, error) {

	mibDb := MibDb{
		items: []MibDbEntry{},
	}

	// the first element to return is the ONU-Data
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.OnuDataClassID,
		EntityID{0x00, 0x00},
		me.AttributeValueMap{"MibDataSync": 0}, // FIXME this needs to be parametrized before sending the response
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
			"Type":          xgsPonUnitType,
			"NumberOfPorts": 1, // NOTE is this the ANI port? must be
			"SerialNumber":  ToOctets("BBSM-Circuit-Pack-ani", 20),
			"Version":       ToOctets("v0.0.1", 20),
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"VendorId":            "ONF",
			"AdministrativeState": 0,
			"OperationalState":    0,
			"BridgedOrIpInd":      0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"EquipmentId":                 ToOctets("BBSM-Circuit-Pack", 20),
			"CardConfiguration":           0,
			"TotalTContBufferNumber":      8,
			"TotalPriorityQueueNumber":    8,
			"TotalTrafficSchedulerNumber": 0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"PowerShedOverride": uint32(0),
		},
	})

	// ANI-G
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.AniGClassID,
		EntityID{tcontSlotId, aniGId},
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

	// circuitPack Ethernet
	// NOTE the circuit pack is divided in multiple messages as too big to fit in a single one
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
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"VendorId":            "ONF",
			"AdministrativeState": 0,
			"OperationalState":    0,
			"BridgedOrIpInd":      0,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"EquipmentId":                 ToOctets("BBSM-Circuit-Pack", 20),
			"CardConfiguration":           0,
			"TotalTContBufferNumber":      8,
			"TotalPriorityQueueNumber":    8,
			"TotalTrafficSchedulerNumber": 16,
		},
	})
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.CircuitPackClassID,
		circuitPackEntityID,
		me.AttributeValueMap{
			"PowerShedOverride": uint32(0),
		},
	})

	// PPTP and UNI-Gs
	// NOTE this are dependent on the number of UNI this ONU supports
	// Through an identical ID, the UNI-G ME is implicitly linked to an instance of a PPTP
	for i := 1; i <= uniPortCount; i++ {
		uniEntityId := GenerateUniPortEntityId(uint32(i))

		mibDb.items = append(mibDb.items, MibDbEntry{
			me.PhysicalPathTerminationPointEthernetUniClassID,
			uniEntityId,
			me.AttributeValueMap{
				"ExpectedType":                  0,
				"SensedType":                    ethernetUnitType,
				"AutoDetectionConfiguration":    0,
				"EthernetLoopbackConfiguration": 0,
				"AdministrativeState":           0,
				"OperationalState":              0,
				"ConfigurationInd":              3,
				"MaxFrameSize":                  1518,
				"DteOrDceInd":                   0,
				"PauseTime":                     0,
				"BridgedOrIpInd":                2,
				"Arc":                           0,
				"ArcInterval":                   0,
				"PppoeFilter":                   0,
				"PowerControl":                  0,
			},
		})

		mibDb.items = append(mibDb.items, MibDbEntry{
			me.UniGClassID,
			uniEntityId,
			me.AttributeValueMap{
				"AdministrativeState":         0,
				"Deprecated":                  0,
				"ManagementCapability":        0,
				"NonOmciManagementIdentifier": 0,
				"RelayAgentOptions":           0,
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
					"QueueConfigurationOption":                            0,
					"MaximumQueueSize":                                    100,
					"AllocatedQueueSize":                                  100,
					"DiscardBlockCounterResetInterval":                    0,
					"ThresholdValueForDiscardedBlocksDueToBufferOverflow": 0,
					"RelatedPort":                                         relatedPort.ToUint32(),
					"TrafficSchedulerPointer":                             0, //it was hardcoded to 0x0108 in the current implementation
					"Weight":                                              1,
					"BackPressureOperation":                               1,
					"BackPressureTime":                                    0,
					"BackPressureOccurQueueThreshold":                     0,
					"BackPressureClearQueueThreshold":                     0,
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
				"AllocId": 65535,
			},
		})

		tsEntityId := EntityID{cardHolderSlotID, byte(i)}
		mibDb.items = append(mibDb.items, MibDbEntry{
			me.TrafficSchedulerClassID,
			tsEntityId, //was not reported in the original implementation
			me.AttributeValueMap{
				"TContPointer":            tcontEntityId.ToUint16(), // was hardcoded to a non-existing t-cont
				"TrafficSchedulerPointer": 0,
				"Policy":                  02,
				"PriorityWeight":          0,
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
					"QueueConfigurationOption":                            0,
					"MaximumQueueSize":                                    100,
					"AllocatedQueueSize":                                  100,
					"DiscardBlockCounterResetInterval":                    0,
					"ThresholdValueForDiscardedBlocksDueToBufferOverflow": 0,
					"RelatedPort":                                         relatedPort.ToUint32(),
					"TrafficSchedulerPointer":                             tsEntityId.ToUint16(), //it was hardcoded to 0x0108 in the current implementation
					"Weight":                                              1,
					"BackPressureOperation":                               1,
					"BackPressureTime":                                    0,
					"BackPressureOccurQueueThreshold":                     0,
					"BackPressureClearQueueThreshold":                     0,
				},
			})
		}
	}

	// ONU-2g
	mibDb.items = append(mibDb.items, MibDbEntry{
		me.Onu2GClassID,
		EntityID{0x00, 0x00},
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

	mibDb.NumberOfCommands = uint16(len(mibDb.items))

	return &mibDb, nil
}
