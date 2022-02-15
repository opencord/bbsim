/*
 * Copyright (c) 2018 - present.  Boling Consulting Solutions (bcsw.net)
 * Copyright 2020-present Open Networking Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*
 * NOTE: This file was generated, manual edits will be overwritten!
 *
 * Generated by 'goCodeGenerator.py':
 *              https://github.com/cboling/OMCI-parser/README.md
 */

package generated

import "github.com/deckarep/golang-set"

// MacBridgePortPerformanceMonitoringHistoryDataClassID is the 16-bit ID for the OMCI
// Managed entity MAC bridge port performance monitoring history data
const MacBridgePortPerformanceMonitoringHistoryDataClassID = ClassID(52) // 0x0034

var macbridgeportperformancemonitoringhistorydataBME *ManagedEntityDefinition

// MacBridgePortPerformanceMonitoringHistoryData (Class ID: #52 / 0x0034)
//	This ME collects PM data associated with a MAC bridge port. Instances of this ME are created and
//	deleted by the OLT.
//
//	For a complete discussion of generic PM architecture, refer to clause I.4.
//
//	Relationships
//		An instance of this ME is associated with an instance of a MAC bridge port configuration data
//		ME.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. Through an identical ID, this ME is
//			implicitly linked to an instance of the MAC bridge port configuration data ME. (R, setbycreate)
//			(mandatory) (2-bytes)
//
//		Interval End Time
//			This attribute identifies the most recently finished 15-min interval. (R) (mandatory) (1-byte)
//
//		Threshold Data 1_2 Id
//			Threshold data 1/2 ID: This attribute points to an instance of the threshold data 1 ME that
//			contains PM threshold values. Since no threshold value attribute number exceeds 7, a threshold
//			data 2 ME is optional. (R,-W, setbycreate) (mandatory) (2-bytes)
//
//		Forwarded Frame Counter
//			This attribute counts frames transmitted successfully on this port. (R) (mandatory) (4-bytes)
//
//		Delay Exceeded Discard Counter
//			This attribute counts frames discarded on this port because transmission was delayed. (R)
//			(mandatory) (4-bytes)
//
//		Maximum Transmission Unit Mtu Exceeded Discard Counter
//			Maximum transmission unit (MTU) exceeded discard counter: This attribute counts frames discarded
//			on this port because the MTU was exceeded. (R) (mandatory) (4-bytes)
//
//		Received Frame Counter
//			This attribute counts frames received on this port. (R) (mandatory) (4-bytes)
//
//		Received And Discarded Counter
//			This attribute counts frames received on this port that were discarded due to errors. (R)
//			(mandatory) (4-bytes)
//
type MacBridgePortPerformanceMonitoringHistoryData struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const MacBridgePortPerformanceMonitoringHistoryData_IntervalEndTime = "IntervalEndTime"
const MacBridgePortPerformanceMonitoringHistoryData_ThresholdData12Id = "ThresholdData12Id"
const MacBridgePortPerformanceMonitoringHistoryData_ForwardedFrameCounter = "ForwardedFrameCounter"
const MacBridgePortPerformanceMonitoringHistoryData_DelayExceededDiscardCounter = "DelayExceededDiscardCounter"
const MacBridgePortPerformanceMonitoringHistoryData_MaximumTransmissionUnitMtuExceededDiscardCounter = "MaximumTransmissionUnitMtuExceededDiscardCounter"
const MacBridgePortPerformanceMonitoringHistoryData_ReceivedFrameCounter = "ReceivedFrameCounter"
const MacBridgePortPerformanceMonitoringHistoryData_ReceivedAndDiscardedCounter = "ReceivedAndDiscardedCounter"

func init() {
	macbridgeportperformancemonitoringhistorydataBME = &ManagedEntityDefinition{
		Name:    "MacBridgePortPerformanceMonitoringHistoryData",
		ClassID: MacBridgePortPerformanceMonitoringHistoryDataClassID,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
			GetCurrentData,
		),
		AllowedAttributeMask: 0xfe00,
		AttributeDefinitions: AttributeDefinitionMap{
			0: Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1: ByteField(MacBridgePortPerformanceMonitoringHistoryData_IntervalEndTime, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read), false, false, false, 1),
			2: Uint16Field(MacBridgePortPerformanceMonitoringHistoryData_ThresholdData12Id, PointerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3: Uint32Field(MacBridgePortPerformanceMonitoringHistoryData_ForwardedFrameCounter, CounterAttributeType, 0x2000, 0, mapset.NewSetWith(Read), false, false, false, 3),
			4: Uint32Field(MacBridgePortPerformanceMonitoringHistoryData_DelayExceededDiscardCounter, CounterAttributeType, 0x1000, 0, mapset.NewSetWith(Read), false, false, false, 4),
			5: Uint32Field(MacBridgePortPerformanceMonitoringHistoryData_MaximumTransmissionUnitMtuExceededDiscardCounter, CounterAttributeType, 0x0800, 0, mapset.NewSetWith(Read), false, false, false, 5),
			6: Uint32Field(MacBridgePortPerformanceMonitoringHistoryData_ReceivedFrameCounter, CounterAttributeType, 0x0400, 0, mapset.NewSetWith(Read), false, false, false, 6),
			7: Uint32Field(MacBridgePortPerformanceMonitoringHistoryData_ReceivedAndDiscardedCounter, CounterAttributeType, 0x0200, 0, mapset.NewSetWith(Read), false, false, false, 7),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
		Alarms: AlarmMap{
			1: "Delay exceeded discard",
			2: "MTU exceeded discard",
			4: "Received and discarded",
		},
	}
}

// NewMacBridgePortPerformanceMonitoringHistoryData (class ID 52) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewMacBridgePortPerformanceMonitoringHistoryData(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*macbridgeportperformancemonitoringhistorydataBME, params...)
}