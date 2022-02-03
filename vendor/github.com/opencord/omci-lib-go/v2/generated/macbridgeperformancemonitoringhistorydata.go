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

// MacBridgePerformanceMonitoringHistoryDataClassID is the 16-bit ID for the OMCI
// Managed entity MAC bridge performance monitoring history data
const MacBridgePerformanceMonitoringHistoryDataClassID = ClassID(51) // 0x0033

var macbridgeperformancemonitoringhistorydataBME *ManagedEntityDefinition

// MacBridgePerformanceMonitoringHistoryData (Class ID: #51 / 0x0033)
//	This ME collects PM data associated with a MAC bridge. Instances of this ME are created and
//	deleted by the OLT.
//
//	For a complete discussion of generic PM architecture, refer to clause I.4.
//
//	Relationships
//		This ME is associated with an instance of a MAC bridge service profile.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. Through an identical ID, this ME is
//			implicitly linked to an instance of the MAC bridge service profile. (R, setbycreate) (mandatory)
//			(2-bytes)
//
//		Interval End Time
//			This attribute identifies the most recently finished 15-min interval. (R) (mandatory) (1-byte)
//
//		Threshold Data 1_2 Id
//			Threshold data 1/2 ID: This attribute points to an instance of the threshold data 1 ME that
//			contains PM threshold values. Since no threshold value attribute number exceeds 7, a threshold
//			data 2 ME is optional. Since no threshold value attribute number exceeds 7, a threshold data 2
//			ME is optional. (R,-W, setbycreate) (mandatory) (2-bytes)
//
//		Bridge Learning Entry Discard Count
//			This attribute counts forwarding database entries that have been or would have been learned, but
//			were discarded or replaced due to a lack of space in the database table. When used with the MAC
//			learning depth attribute of the MAC bridge service profile, the bridge learning entry discard
//			count may be particularly useful in detecting MAC spoofing attempts. (R) (mandatory) (4-bytes)
//
type MacBridgePerformanceMonitoringHistoryData struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const MacBridgePerformanceMonitoringHistoryData_IntervalEndTime = "IntervalEndTime"
const MacBridgePerformanceMonitoringHistoryData_ThresholdData12Id = "ThresholdData12Id"
const MacBridgePerformanceMonitoringHistoryData_BridgeLearningEntryDiscardCount = "BridgeLearningEntryDiscardCount"

func init() {
	macbridgeperformancemonitoringhistorydataBME = &ManagedEntityDefinition{
		Name:    "MacBridgePerformanceMonitoringHistoryData",
		ClassID: MacBridgePerformanceMonitoringHistoryDataClassID,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
			GetCurrentData,
		),
		AllowedAttributeMask: 0xe000,
		AttributeDefinitions: AttributeDefinitionMap{
			0: Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1: ByteField(MacBridgePerformanceMonitoringHistoryData_IntervalEndTime, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read), false, false, false, 1),
			2: Uint16Field(MacBridgePerformanceMonitoringHistoryData_ThresholdData12Id, PointerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3: Uint32Field(MacBridgePerformanceMonitoringHistoryData_BridgeLearningEntryDiscardCount, CounterAttributeType, 0x2000, 0, mapset.NewSetWith(Read), false, false, false, 3),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
		Alarms: AlarmMap{
			0: "Bridge learning entry discard",
		},
	}
}

// NewMacBridgePerformanceMonitoringHistoryData (class ID 51) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewMacBridgePerformanceMonitoringHistoryData(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*macbridgeperformancemonitoringhistorydataBME, params...)
}
