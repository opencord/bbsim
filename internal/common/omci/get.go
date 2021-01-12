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
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	log "github.com/sirupsen/logrus"
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

func CreateGetResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseGetRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":    msgObj.EntityClass,
		"EntityInstance": msgObj.EntityInstance,
		"AttributeMask":  fmt.Sprintf("%x", msgObj.AttributeMask),
	}).Debug("recevied-omci-get-request")

	var response *omci.GetResponse
	switch msgObj.EntityClass {
	case me.Onu2GClassID:
		response = CreateOnu2gResponse(msgObj.AttributeMask)
	case me.OnuGClassID:
		response = CreateOnugResponse(msgObj.AttributeMask)
	case me.SoftwareImageClassID:
		response = CreateSoftwareImageResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	case me.IpHostConfigDataClassID:
		response = CreateIpHostResponse(msgObj.AttributeMask, msgObj.EntityInstance)
	default:
		log.Warnf("don't know how to handle me class: %s", msgObj.EntityClass.String())
		return nil, nil
	}

	pkt, err := serialize(omci.GetResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err":  err,
			"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		}).Error("Cannot serialize Onu2gResponse")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Info("GetResponse")

	return pkt, nil
}

func CreateOnu2gResponseOld() *omci.GetResponse {
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.Onu2GClassID,
			EntityInstance: 0,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": 0,
			"EquipmentId":     toOctets("BBSM", 20),
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
		Result:        me.Success,
		AttributeMask: 0xE000,
	}
}

func CreateOnu2gResponse(attributeMask uint16) *omci.GetResponse {

	managedEntity, meErr := me.NewOnu2G(me.ParamData{Attributes: me.AttributeValueMap{
		"ManagedEntityId": 0,
		"EquipmentId":     toOctets("BBSM", 20),
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
	}})

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

func CreateOnugResponse(attributeMask uint16) *omci.GetResponse {
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.OnuGClassID,
			EntityInstance: 0,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId":         0,
			"VendorId":                toOctets("BBSM", 4),
			"Version":                 toOctets("v0.0.1", 14),
			"SerialNumber":            toOctetsB("QkJTTQAKAAE="),
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
		Result:        me.Success,
		AttributeMask: attributeMask,
	}
}

func CreateSoftwareImageResponse(attributeMask uint16, entityInstance uint16) *omci.GetResponse {
	// NOTE that we need send the response for the correct ME Instance or the adapter won't process it
	return &omci.GetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.SoftwareImageClassID,
			EntityInstance: entityInstance,
		},
		Attributes: me.AttributeValueMap{
			"ManagedEntityId": 0,
			"Version":         toOctets("v0.0.1", 14),
			"IsCommitted":     1,
			"IsActive":        1,
			"IsValid":         1,
			"ProductCode":     toOctets("product-code", 25),
			"ImageHash":       toOctets("broadband-sim", 16),
		},
		Result:        me.Success,
		AttributeMask: attributeMask,
	}
}

func CreateIpHostResponse(attributeMask uint16, entityInstance uint16) *omci.GetResponse {
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

func toOctetsB(str string) []byte {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(fmt.Sprintf("Invalid Base-64 string: '%v'", str))
	}
	return data
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
