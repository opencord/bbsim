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
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/google/gopacket"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	log "github.com/sirupsen/logrus"
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

func CreateGetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI, onuSn *openolt.SerialNumber, mds uint8,
	activeImageEntityId uint16, committedImageEntityId uint16, standbyImageVersion string, activeImageVersion string,
	committedImageVersion string, onuDown bool) ([]byte, error) {
	msgObj, err := ParseGetRequest(omciPkt)
	if err != nil {
		return nil, err
	}
	omciLogger.WithFields(log.Fields{
		"DeviceIdent":    omciMsg.DeviceIdentifier,
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  fmt.Sprintf("%x", msgObj.AttributeMask),
	}).Trace("received-omci-get-request")

	var response *omci.GetResponse

	isExtended := false
	if omciMsg.DeviceIdentifier == omci.ExtendedIdent {
		isExtended = true
	}
	switch msgObj.EntityClass {
	case me.Onu2GClassID:
		response = createOnu2gResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.OnuGClassID:
		response = createOnugResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance, onuSn)
	case me.SoftwareImageClassID:
		response = createSoftwareImageResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance,
			activeImageEntityId, committedImageEntityId, standbyImageVersion, activeImageVersion, committedImageVersion)
	case me.IpHostConfigDataClassID:
		response = createIpHostResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.VoipConfigDataClassID:
		response = createVoipConfigDataResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.UniGClassID:
		response = createUnigResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance, onuDown)
	case me.PhysicalPathTerminationPointEthernetUniClassID:
		response = createPptpEthernetResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance, onuDown)
	case me.PhysicalPathTerminationPointPotsUniClassID:
		response = createPptpPotsResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance, onuDown)
	case me.AniGClassID:
		response = createAnigResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.OnuDataClassID:
		response = createOnuDataResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance, mds)
	case me.EthernetFramePerformanceMonitoringHistoryDataUpstreamClassID:
		response = createEthernetFramePerformanceMonitoringHistoryDataUpstreamResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID:
		response = createEthernetFramePerformanceMonitoringHistoryDataDownstreamResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.EthernetPerformanceMonitoringHistoryDataClassID:
		response = createEthernetPerformanceMonitoringHistoryDataResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.FecPerformanceMonitoringHistoryDataClassID:
		response = createFecPerformanceMonitoringHistoryDataResponse(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.GemPortNetworkCtpPerformanceMonitoringHistoryDataClassID:
		response = createGemPortNetworkCtpPerformanceMonitoringHistoryData(isExtended, msgObj.AttributeMask, msgObj.EntityInstance)
	case me.EthernetFrameExtendedPmClassID,
		me.EthernetFrameExtendedPm64BitClassID:
		response = createEthernetFrameExtendedPmGetResponse(isExtended, msgObj.EntityClass, msgObj.AttributeMask, msgObj.EntityInstance)
	default:
		omciLogger.WithFields(log.Fields{
			"EntityClass":    msgObj.EntityClass,
			"EntityInstance": msgObj.EntityInstance,
			"AttributeMask":  fmt.Sprintf("%x", msgObj.AttributeMask),
		}).Warnf("do-not-know-how-to-handle-get-request-for-me-class")
		return nil, nil
	}
	omciLayer := &omci.OMCI{
		TransactionID:    omciMsg.TransactionID,
		MessageType:      omci.GetResponseType,
		DeviceIdentifier: omciMsg.DeviceIdentifier,
	}
	var options gopacket.SerializeOptions
	options.FixLengths = true

	buffer := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(buffer, options, omciLayer, response)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err":  err,
			"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		}).Error("cannot-Serialize-GetResponse")
		return nil, err
	}
	pkt := buffer.Bytes()

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Trace("omci-get-response")

	return pkt, nil
}

func createOnu2gResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {

	managedEntity, meErr := me.NewOnu2G(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:   entityID,
			me.Onu2G_EquipmentId: ToOctets("12345123451234512345", 20),
			me.Onu2G_OpticalNetworkUnitManagementAndControlChannelOmccVersion: common.Config.BBSim.OmccVersion,
			me.Onu2G_VendorProductCode:                                        0,
			me.Onu2G_SecurityCapability:                                       1,
			me.Onu2G_SecurityMode:                                             1,
			me.Onu2G_TotalPriorityQueueNumber:                                 1,
			me.Onu2G_TotalTrafficSchedulerNumber:                              1,
			me.Onu2G_Deprecated:                                               1,
			me.Onu2G_TotalGemPortIdNumber:                                     32,
			me.Onu2G_Sysuptime:                                                319389947, // NOTE need to be smarter?
			me.Onu2G_ConnectivityCapability:                                   127,
			me.Onu2G_CurrentConnectivityMode:                                  5,
			me.Onu2G_QualityOfServiceQosConfigurationFlexibility:              48,
			me.Onu2G_PriorityQueueScaleFactor:                                 1,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewOnu2G %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.Onu2GClassID,
			EntityInstance: entityID,
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createOnugResponse(isExtended bool, attributeMask uint16, entityID uint16, onuSn *openolt.SerialNumber) *omci.GetResponse {

	managedEntity, meErr := me.NewOnuG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:              entityID,
			me.OnuG_VendorId:                ToOctets("BBSM", 4),
			me.OnuG_Version:                 ToOctets("v0.0.1", 14),
			me.OnuG_SerialNumber:            append(onuSn.VendorId, onuSn.VendorSpecific...),
			me.OnuG_TrafficManagementOption: 0,
			me.OnuG_Deprecated:              0,
			me.OnuG_BatteryBackup:           0,
			me.OnuG_AdministrativeState:     0,
			me.OnuG_OperationalState:        0,
			me.OnuG_OnuSurvivalTime:         10,
			me.OnuG_LogicalOnuId:            ToOctets("BBSM", 24),
			me.OnuG_LogicalPassword:         ToOctets("BBSM", 12),
			me.OnuG_CredentialsStatus:       0,
			me.OnuG_ExtendedTcLayerOptions:  0,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewOnu2G %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.OnuGClassID,
			EntityInstance: entityID,
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createSoftwareImageResponse(isExtended bool, attributeMask uint16, entityInstance uint16, activeImageEntityId uint16,
	committedImageEntityId uint16, standbyImageVersion string, activeImageVersion string, committedImageVersion string) *omci.GetResponse {

	omciLogger.WithFields(log.Fields{
		"EntityInstance": entityInstance,
		"AttributeMask":  attributeMask,
	}).Trace("received-get-software-image-request")

	// Only one image can be active and committed
	committed := 0
	active := 0
	version := standbyImageVersion
	if entityInstance == activeImageEntityId {
		active = 1
		version = activeImageVersion
	}
	if entityInstance == committedImageEntityId {
		committed = 1
		version = committedImageVersion
	}

	imageHash, err := hex.DecodeString(hex.EncodeToString([]byte(version)))
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"entityId":  entityInstance,
			"active":    active,
			"committed": committed,
			"err":       err,
		}).Error("cannot-generate-image-hash")
	}

	// NOTE that we need send the response for the correct ME Instance or the adapter won't process it
	res := &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.SoftwareImageClassID,
			EntityInstance: entityInstance,
			Extended:       isExtended,
		},
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:           0,
			me.SoftwareImage_Version:     ToOctets(version, 14),
			me.SoftwareImage_IsCommitted: committed,
			me.SoftwareImage_IsActive:    active,
			me.SoftwareImage_IsValid:     1,
			me.SoftwareImage_ProductCode: ToOctets("BBSIM-ONU", 25),
			me.SoftwareImage_ImageHash:   imageHash,
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

func createIpHostResponse(isExtended bool, attributeMask uint16, entityInstance uint16) *omci.GetResponse {
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.IpHostConfigDataClassID,
			EntityInstance: entityInstance,
			Extended:       isExtended,
		},
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:             0,
			me.IpHostConfigData_MacAddress: ToOctets("aabbcc", 6),
		},
		Result:        me.Success,
		AttributeMask: attributeMask,
	}
}

func createVoipConfigDataResponse(isExtended bool, attributeMask uint16, entityInstance uint16) *omci.GetResponse {
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.VoipConfigDataClassID,
			EntityInstance: entityInstance,
			Extended:       isExtended,
		},
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: 0,
			me.VoipConfigData_AvailableSignallingProtocols:      1,
			me.VoipConfigData_SignallingProtocolUsed:            1,
			me.VoipConfigData_AvailableVoipConfigurationMethods: 1,
			me.VoipConfigData_VoipConfigurationMethodUsed:       1,
			me.VoipConfigData_VoipConfigurationAddressPointer:   0xFFFF,
			me.VoipConfigData_VoipConfigurationState:            0,
			me.VoipConfigData_RetrieveProfile:                   0,
			me.VoipConfigData_ProfileVersion:                    0,
		},
		Result:        me.Success,
		AttributeMask: attributeMask,
	}
}

func createUnigResponse(isExtended bool, attributeMask uint16, entityID uint16, onuDown bool) *omci.GetResponse {
	// Valid values for uni_admin_state are 0 (unlocks) and 1 (locks)
	omciAdminState := 1
	if !onuDown {
		omciAdminState = 0
	}
	managedEntity, meErr := me.NewUniG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:                  entityID,
			me.UniG_Deprecated:                  0,
			me.UniG_AdministrativeState:         omciAdminState,
			me.UniG_ManagementCapability:        0,
			me.UniG_NonOmciManagementIdentifier: 1,
			me.UniG_RelayAgentOptions:           1,
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createPptpEthernetResponse(isExtended bool, attributeMask uint16, entityID uint16, onuDown bool) *omci.GetResponse {
	// Valid values for oper_state are 0 (enabled) and 1 (disabled)
	// Valid values for uni_admin_state are 0 (unlocks) and 1 (locks)
	onuAdminState := 1
	if !onuDown {
		onuAdminState = 0
	}
	onuOperState := onuAdminState // For now make the assumption that oper state reflects the admin state
	managedEntity, meErr := me.NewPhysicalPathTerminationPointEthernetUni(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.PhysicalPathTerminationPointEthernetUni_ExpectedType:                  0,
			me.PhysicalPathTerminationPointEthernetUni_SensedType:                    ethernetUnitType,
			me.PhysicalPathTerminationPointEthernetUni_AutoDetectionConfiguration:    0,
			me.PhysicalPathTerminationPointEthernetUni_EthernetLoopbackConfiguration: 0,
			me.PhysicalPathTerminationPointEthernetUni_AdministrativeState:           onuAdminState,
			me.PhysicalPathTerminationPointEthernetUni_OperationalState:              onuOperState,
			me.PhysicalPathTerminationPointEthernetUni_ConfigurationInd:              3,
			me.PhysicalPathTerminationPointEthernetUni_MaxFrameSize:                  0,
			me.PhysicalPathTerminationPointEthernetUni_DteOrDceInd:                   0,
			me.PhysicalPathTerminationPointEthernetUni_PauseTime:                     0,
			me.PhysicalPathTerminationPointEthernetUni_BridgedOrIpInd:                0,
			me.PhysicalPathTerminationPointEthernetUni_Arc:                           0,
			me.PhysicalPathTerminationPointEthernetUni_ArcInterval:                   0,
			me.PhysicalPathTerminationPointEthernetUni_PppoeFilter:                   0,
			me.PhysicalPathTerminationPointEthernetUni_PowerControl:                  0,
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createPptpPotsResponse(isExtended bool, attributeMask uint16, entityID uint16, onuDown bool) *omci.GetResponse {
	// Valid values for oper_state are 0 (enabled) and 1 (disabled)
	// Valid values for uni_admin_state are 0 (unlocks) and 1 (locks)
	onuAdminState := 1
	if !onuDown {
		onuAdminState = 0
	}
	onuOperState := onuAdminState // For now make the assumption that oper state reflects the admin state
	managedEntity, meErr := me.NewPhysicalPathTerminationPointPotsUni(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.PhysicalPathTerminationPointPotsUni_AdministrativeState: onuAdminState,
			me.PhysicalPathTerminationPointPotsUni_Deprecated:          0,
			me.PhysicalPathTerminationPointPotsUni_Arc:                 0,
			me.PhysicalPathTerminationPointPotsUni_ArcInterval:         0,
			me.PhysicalPathTerminationPointPotsUni_Impedance:           0,
			me.PhysicalPathTerminationPointPotsUni_TransmissionPath:    0,
			me.PhysicalPathTerminationPointPotsUni_RxGain:              0,
			me.PhysicalPathTerminationPointPotsUni_TxGain:              0,
			me.PhysicalPathTerminationPointPotsUni_OperationalState:    onuOperState,
			me.PhysicalPathTerminationPointPotsUni_HookState:           0,
			me.PhysicalPathTerminationPointPotsUni_PotsHoldoverTime:    0,
			me.PhysicalPathTerminationPointPotsUni_NominalFeedVoltage:  0,
			me.PhysicalPathTerminationPointPotsUni_LossOfSoftswitch:    0,
		},
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewPhysicalPathTerminationPointPotsUni %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.PhysicalPathTerminationPointPotsUniClassID,
			EntityInstance: entityID,
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetFramePerformanceMonitoringHistoryDataUpstreamResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetFramePerformanceMonitoringHistoryDataUpstream(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_IntervalEndTime:         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_ThresholdData12Id:       0,
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_DropEvents:              rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Octets:                  rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets:                 rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_BroadcastPackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_MulticastPackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_CrcErroredPackets:       rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_UndersizePackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_OversizePackets:         rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets64Octets:         rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets65To127Octets:    rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets128To255Octets:   rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets256To511Octets:   rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets512To1023Octets:  rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataUpstream_Packets1024To1518Octets: rand.Intn(100),
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetFramePerformanceMonitoringHistoryDataDownstreamResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetFramePerformanceMonitoringHistoryDataDownstream(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_IntervalEndTime:         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_ThresholdData12Id:       0,
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_DropEvents:              rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Octets:                  rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets:                 rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_BroadcastPackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_MulticastPackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_CrcErroredPackets:       rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_UndersizePackets:        rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_OversizePackets:         rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets64Octets:         rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets65To127Octets:    rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets128To255Octets:   rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets256To511Octets:   rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets512To1023Octets:  rand.Intn(100),
			me.EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets1024To1518Octets: rand.Intn(100),
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetPerformanceMonitoringHistoryDataResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewEthernetPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.EthernetPerformanceMonitoringHistoryData_IntervalEndTime:                 0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			me.EthernetPerformanceMonitoringHistoryData_ThresholdData12Id:               0,
			me.EthernetPerformanceMonitoringHistoryData_FcsErrors:                       rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_ExcessiveCollisionCounter:       rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_LateCollisionCounter:            rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_FramesTooLong:                   rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_BufferOverflowsOnReceive:        rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_BufferOverflowsOnTransmit:       rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_SingleCollisionFrameCounter:     rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_MultipleCollisionsFrameCounter:  rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_SqeCounter:                      rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_DeferredTransmissionCounter:     rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_InternalMacTransmitErrorCounter: rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_CarrierSenseErrorCounter:        rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_AlignmentErrorCounter:           rand.Intn(100),
			me.EthernetPerformanceMonitoringHistoryData_InternalMacReceiveErrorCounter:  rand.Intn(100),
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createFecPerformanceMonitoringHistoryDataResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewFecPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.FecPerformanceMonitoringHistoryData_IntervalEndTime:        0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			me.FecPerformanceMonitoringHistoryData_ThresholdData12Id:      0,
			me.FecPerformanceMonitoringHistoryData_CorrectedBytes:         rand.Intn(100),
			me.FecPerformanceMonitoringHistoryData_CorrectedCodeWords:     rand.Intn(100),
			me.FecPerformanceMonitoringHistoryData_UncorrectableCodeWords: rand.Intn(100),
			me.FecPerformanceMonitoringHistoryData_TotalCodeWords:         rand.Intn(100),
			me.FecPerformanceMonitoringHistoryData_FecSeconds:             rand.Intn(100),
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createGemPortNetworkCtpPerformanceMonitoringHistoryData(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewGemPortNetworkCtpPerformanceMonitoringHistoryData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID: entityID,
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_IntervalEndTime:         0, // This ideally should increment by 1 every collection interval, but staying 0 for simulation is Ok for now.
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_ThresholdData12Id:       0,
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_TransmittedGemFrames:    rand.Intn(100),
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_ReceivedGemFrames:       rand.Intn(100),
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_ReceivedPayloadBytes:    rand.Intn(100),
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_TransmittedPayloadBytes: rand.Intn(100),
			me.GemPortNetworkCtpPerformanceMonitoringHistoryData_EncryptionKeyErrors:     rand.Intn(100),
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createOnuDataResponse(isExtended bool, attributeMask uint16, entityID uint16, mds uint8) *omci.GetResponse {
	managedEntity, meErr := me.NewOnuData(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:     entityID,
			me.OnuData_MibDataSync: mds,
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createAnigResponse(isExtended bool, attributeMask uint16, entityID uint16) *omci.GetResponse {
	managedEntity, meErr := me.NewAniG(me.ParamData{
		EntityID: entityID,
		Attributes: me.AttributeValueMap{
			me.ManagedEntityID:                  entityID,
			me.AniG_SrIndication:                0,
			me.AniG_TotalTcontNumber:            0,
			me.AniG_GemBlockLength:              0,
			me.AniG_PiggybackDbaReporting:       0,
			me.AniG_Deprecated:                  0,
			me.AniG_SignalFailThreshold:         0,
			me.AniG_SignalDegradeThreshold:      0,
			me.AniG_Arc:                         0,
			me.AniG_ArcInterval:                 0,
			me.AniG_OpticalSignalLevel:          rand.Intn(16000), // generate some random power level than defaulting to 0
			me.AniG_LowerOpticalThreshold:       0,
			me.AniG_UpperOpticalThreshold:       0,
			me.AniG_OnuResponseTime:             0,
			me.AniG_TransmitOpticalLevel:        rand.Intn(16000), // generate some random power level than defaulting to 0
			me.AniG_LowerTransmitPowerThreshold: 0,
			me.AniG_UpperTransmitPowerThreshold: 0,
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
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func createEthernetFrameExtendedPmGetResponse(isExtended bool, meClass me.ClassID, attributeMask uint16, entityID uint16) *omci.GetResponse {

	callback := me.NewEthernetFrameExtendedPm
	if meClass != me.EthernetFrameExtendedPmClassID {
		callback = me.NewEthernetFrameExtendedPm64Bit
	}

	//The names of these attributes are left as strings
	//rather than constants of a particular ME because
	//they can be used with both the MEs in the lines above
	attr := me.AttributeValueMap{
		me.ManagedEntityID:       entityID,
		"DropEvents":             100,
		"Octets":                 101,
		"Frames":                 102,
		"BroadcastFrames":        103,
		"MulticastFrames":        104,
		"CrcErroredFrames":       105,
		"UndersizeFrames":        106,
		"OversizeFrames":         107,
		"Frames64Octets":         108,
		"Frames65To127Octets":    109,
		"Frames128To255Octets":   110,
		"Frames256To511Octets":   111,
		"Frames512To1023Octets":  112,
		"Frames1024To1518Octets": 113,
	}
	managedEntity, meErr := callback(me.ParamData{
		EntityID:   entityID,
		Attributes: attr,
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("NewEthernetFrameExtendedPm %v", meErr.Error())
		return nil
	}

	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    meClass,
			EntityInstance: entityID,
			Extended:       isExtended,
		},
		Attributes:    managedEntity.GetAttributeValueMap(),
		AttributeMask: attributeMask,
		Result:        me.Success,
	}
}

func ToOctets(str string, size int) []byte {
	asciiBytes := []byte(str)

	if len(asciiBytes) < size {
		missing := size - len(asciiBytes)
		for i := 0; i < missing; i++ {
			asciiBytes = append(asciiBytes, []byte{0x00}[0])
		}
	}
	return asciiBytes
}
