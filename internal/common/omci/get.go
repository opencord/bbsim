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
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
)

func ParseGetRequest(omciPkt gopacket.Packet) (*omci.GetRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeGetRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.GetRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateGetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI, onuSn *openolt.SerialNumber, mds uint8, activeImageEntityId uint16, committedImageEntityId uint16) ([]byte, error) {

	msgObj, err := ParseGetRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  fmt.Sprintf("%x", msgObj.AttributeMask),
	}).Trace("received-omci-get-request")

	var response *omci.GetResponse
	switch msgObj.EntityClass {
	case me.Onu2GClassID:
		response = createOnu2gResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.OnuGClassID:
		response = createOnugResponse(msgObj.AttributeMask, msgObj.EntityInstance, onuSn)
	case me.SoftwareImageClassID:
		response = createSoftwareImageResponse(msgObj.AttributeMask, msgObj.EntityInstance, activeImageEntityId, committedImageEntityId)
	case me.IpHostConfigDataClassID:
		response = createIpHostResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.UniGClassID:
		response = createUnigResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.PhysicalPathTerminationPointEthernetUniClassID:
		response = createPptpResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.AniGClassID:
		response = createAnigResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.OnuDataClassID:
		response = createOnuDataResponse(msgObj.AttributeMask, msgObj.EntityInstance, mds)
	case me.EthernetFramePerformanceMonitoringHistoryDataUpstreamClassID:
		response = createEthernetFramePerformanceMonitoringHistoryDataUpstreamResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID:
		response = createEthernetFramePerformanceMonitoringHistoryDataDownstreamResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.EthernetPerformanceMonitoringHistoryDataClassID:
		response = createEthernetPerformanceMonitoringHistoryDataResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.FecPerformanceMonitoringHistoryDataClassID:
		response = createFecPerformanceMonitoringHistoryDataResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.GemPortNetworkCtpPerformanceMonitoringHistoryDataClassID:
		response = createGemPortNetworkCtpPerformanceMonitoringHistoryData(msgObj.AttributeMask, msgObj.EntityInstance)
	default:
		omciLogger.WithFields(log.Fields{
			"EntityClass":    msgObj.EntityClass,
			"EntityInstance": msgObj.EntityInstance,
			"AttributeMask":  fmt.Sprintf("%x", msgObj.AttributeMask),
		}).Warnf("do-not-know-how-to-handle-get-request-for-me-class")
		return nil, nil
	}

	pkt, err := Serialize(omci.GetResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err":  err,
			"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		}).Error("cannot-Serialize-GetResponse")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Trace("omci-get-response")

	return pkt, nil
}

func createOnu2gResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {

	managedEntity, meErr := me.NewOnu2G(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": entityID,
			"EquipmentId":     toOctets("12345123451234512345", 20),
			"OpticalNetworkUnitManagementAndControlChannelOmccVersion": 180,
			"VendorProductCode":                           0,
			"SecurityCapability":                          1,
			"SecurityMode":                                1,
			"TotalPriorityQueueNumber":                    1,
			"TotalTrafficSchedulerNumber":                 1,
			"Deprecated":                                  1,
			"TotalGemPortIdNumber":                        32,
			"Sysuptime":                                   319389947, // NOTE need to be smarter?
			"ConnectivityCapability":                      127,
			"CurrentConnectivityMode":                     5,
			"QualityOfServiceQosConfigurationFlexibility": 48,
			"PriorityQueueScaleFactor":                    1,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewOnu2G %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.Onu2GClassID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createOnugResponse(attributeMask uint16, entityID uint16, onuSn *openolt.SerialNumber) *omci.GetResponse {

	managedEntity, meErr := me.NewOnuG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":         entityID,
			"VendorId":                toOctets("BBSM", 4),
			"Version":                 toOctets("v0.0.1", 14),
			"SerialNumber":            append(onuSn.VendorId, onuSn.VendorSpecific...),
			"TrafficManagementOption": 0,
			"Deprecated":              0,
			"BatteryBackup":           0,
			"AdministrativeState":     0,
			"OperationalState":        0,
			"OnuSurvivalTime":         10,
			"LogicalOnuId":            toOctets("BBSM", 24),
			"LogicalPassword":         toOctets("BBSM", 12),
			"CredentialsStatus":       0,
			"ExtendedTcLayerOptions":  0,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewOnu2G %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuGClassID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}

	//return &omci.GetResponse{
	//	MeBasePacket: omci.MeBasePacket{
	//		EntityClass:    me.OnuGClassID,
	//		EntityInstance: entityID,
	//	},
	//	Attributes: me.AttributeValueMap{
	//
	//	},
	//	Result:        me.Success,
	//	AttributeMask: attributeMask,
	//}
}

func createSoftwareImageResponse(attributeMask uint16, entityInstance uint16, activeImageEntityId uint16, committedImageEntityId uint16) *omci.GetResponse {

	omciLogger.WithFields(log.Fields{
		"EntityInstance": entityInstance,
	}).Trace("received-get-software-image-request")

	// Only one image can be active and committed
	committed := 0
	active := 0
	if entityInstance == activeImageEntityId {
		active = 1
	}
	if entityInstance == committedImageEntityId {
		committed = 1
	}

	// NOTE that we need send the response for the correct ME Instance or the adapter won't process it
	res := &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.SoftwareImageClassID,
			EntityInstance: entityInstance,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": 0,
			"Version":         toOctets("00000000000001", 14),
			"IsCommitted":     committed,
			"IsActive":        active,
			"IsValid":         1,
			"ProductCode":     toOctets("product-code", 25),
			"ImageHash":       toOctets("broadband-sim", 16),
		},
		Result:        me.Success,
		AttributeMask: attributeMask,
	}

	omciLogger.WithFields(log.Fields{
		"omciMessage": res,
		"entityId":    entityInstance,
		"active":      active,
		"committed":   committed,
	}).Trace("Reporting SoftwareImage")

	return res
}

func createIpHostResponse(attributeMask uint16, entityInstance uint16) *omci.GetResponse {
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.IpHostConfigDataClassID,
			EntityInstance: entityInstance,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": 0,
			"MacAddress":      toOctets("aabbcc", 6),
		},
		Result:        me.Success,
		AttributeMask: attributeMask,
	}
}

func createUnigResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewUniG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":             entityID,
			"Deprecated":                  0,
			"AdministrativeState":         0,
			"ManagementCapability":        0,
			"NonOmciManagementIdentifier": 1,
			"RelayAgentOptions":           1,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewUniG %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.UniGClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createPptpResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewPhysicalPathTerminationPointEthernetUni(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":               entityID,
			"ExpectedType":                  0,
			"SensedType":                    0,
			"AutoDetectionConfiguration":    0,
			"EthernetLoopbackConfiguration": 0,
			"AdministrativeState":           0,
			"OperationalState":              0,
			"ConfigurationInd":              0,
			"MaxFrameSize":                  0,
			"DteOrDceInd":                   0,
			"PauseTime":                     0,
			"BridgedOrIpInd":                0,
			"Arc":                           0,
			"ArcInterval":                   0,
			"PppoeFilter":                   0,
			"PowerControl":                  0,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewPhysicalPathTerminationPointEthernetUni %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.PhysicalPathTerminationPointEthernetUniClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetFramePerformanceMonitoringHistoryDataUpstreamResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetFramePerformanceMonitoringHistoryDataUpstream(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":         entityID,
			"IntervalEndTime":         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			"ThresholdData12Id":       0,
			"DropEvents":              rand.Intn(100),
			"Octets":                  rand.Intn(100),
			"Packets":                 rand.Intn(100),
			"BroadcastPackets":        rand.Intn(100),
			"MulticastPackets":        rand.Intn(100),
			"CrcErroredPackets":       rand.Intn(100),
			"UndersizePackets":        rand.Intn(100),
			"OversizePackets":         rand.Intn(100),
			"Packets64Octets":         rand.Intn(100),
			"Packets65To127Octets":    rand.Intn(100),
			"Packets128To255Octets":   rand.Intn(100),
			"Packets256To511Octets":   rand.Intn(100),
			"Packets512To1023Octets":  rand.Intn(100),
			"Packets1024To1518Octets": rand.Intn(100),
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewEthernetFramePerformanceMonitoringHistoryDataUpstream %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.EthernetFramePerformanceMonitoringHistoryDataUpstreamClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetFramePerformanceMonitoringHistoryDataDownstreamResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetFramePerformanceMonitoringHistoryDataDownstream(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":         entityID,
			"IntervalEndTime":         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			"ThresholdData12Id":       0,
			"DropEvents":              rand.Intn(100),
			"Octets":                  rand.Intn(100),
			"Packets":                 rand.Intn(100),
			"BroadcastPackets":        rand.Intn(100),
			"MulticastPackets":        rand.Intn(100),
			"CrcErroredPackets":       rand.Intn(100),
			"UndersizePackets":        rand.Intn(100),
			"OversizePackets":         rand.Intn(100),
			"Packets64Octets":         rand.Intn(100),
			"Packets65To127Octets":    rand.Intn(100),
			"Packets128To255Octets":   rand.Intn(100),
			"Packets256To511Octets":   rand.Intn(100),
			"Packets512To1023Octets":  rand.Intn(100),
			"Packets1024To1518Octets": rand.Intn(100),
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewEthernetFramePerformanceMonitoringHistoryDataDownstream %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetPerformanceMonitoringHistoryDataResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":                 entityID,
			"IntervalEndTime":                 0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			"ThresholdData12Id":               0,
			"FcsErrors":                       rand.Intn(100),
			"ExcessiveCollisionCounter":       rand.Intn(100),
			"LateCollisionCounter":            rand.Intn(100),
			"FramesTooLong":                   rand.Intn(100),
			"BufferOverflowsOnReceive":        rand.Intn(100),
			"BufferOverflowsOnTransmit":       rand.Intn(100),
			"SingleCollisionFrameCounter":     rand.Intn(100),
			"MultipleCollisionsFrameCounter":  rand.Intn(100),
			"SqeCounter":                      rand.Intn(100),
			"DeferredTransmissionCounter":     rand.Intn(100),
			"InternalMacTransmitErrorCounter": rand.Intn(100),
			"CarrierSenseErrorCounter":        rand.Intn(100),
			"AlignmentErrorCounter":           rand.Intn(100),
			"InternalMacReceiveErrorCounter":  rand.Intn(100),
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewEthernetPerformanceMonitoringHistoryData %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.EthernetPerformanceMonitoringHistoryDataClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createFecPerformanceMonitoringHistoryDataResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewFecPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":        entityID,
			"IntervalEndTime":        0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			"ThresholdData12Id":      0,
			"CorrectedBytes":         rand.Intn(100),
			"CorrectedCodeWords":     rand.Intn(100),
			"UncorrectableCodeWords": rand.Intn(100),
			"TotalCodeWords":         rand.Intn(100),
			"FecSeconds":             rand.Intn(100),
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewFecPerformanceMonitoringHistoryData %v", meErr.Error())
		return nil
	}

	// FEC History counter fits within single gem payload.
	// No need of the logical we use in other Ethernet History counters or Gem Port History counters

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.FecPerformanceMonitoringHistoryDataClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createGemPortNetworkCtpPerformanceMonitoringHistoryData(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewGemPortNetworkCtpPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":         entityID,
			"IntervalEndTime":         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			"ThresholdData12Id":       0,
			"TransmittedGemFrames":    rand.Intn(100),
			"ReceivedGemFrames":       rand.Intn(100),
			"ReceivedPayloadBytes":    rand.Intn(100),
			"TransmittedPayloadBytes": rand.Intn(100),
			"EncryptionKeyErrors":     rand.Intn(100),
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewGemPortNetworkCtpPerformanceMonitoringHistoryData %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpPerformanceMonitoringHistoryDataClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createOnuDataResponse(attributeMask uint16, entityID uint16, mds uint8) *omci.GetResponse {
	managedEntity, meErr := me.NewOnuData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": entityID,
			"MibDataSync":     mds,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewOnuData %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.OnuDataClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createAnigResponse(attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewAniG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":             entityID,
			"SrIndication":                0,
			"TotalTcontNumber":            0,
			"GemBlockLength":              0,
			"PiggybackDbaReporting":       0,
			"Deprecated":                  0,
			"SignalFailThreshold":         0,
			"SignalDegradeThreshold":      0,
			"Arc":                         0,
			"ArcInterval":                 0,
			"OpticalSignalLevel":          rand.Intn(16000), // generate some random power level than defaulting to 0
			"LowerOpticalThreshold":       0,
			"UpperOpticalThreshold":       0,
			"OnuResponseTime":             0,
			"TransmitOpticalLevel":        rand.Intn(16000), // generate some random power level than defaulting to 0
			"LowerTransmitPowerThreshold": 0,
			"UpperTransmitPowerThreshold": 0,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewAniG %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.AniGClassID,
			EntityInstance: entityID,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func toOctets(str string, size int) []byte {
	asciiBytes := []byte(str)

	if len(asciiBytes) < size {
		missing := size - len(asciiBytes)
		for i := 0; i < missing; i++ {
			asciiBytes = append(asciiBytes, []byte{0x00}[0])
		}
	}
	return asciiBytes
}
