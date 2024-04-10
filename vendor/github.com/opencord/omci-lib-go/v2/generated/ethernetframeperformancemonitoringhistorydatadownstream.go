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

// EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID is the 16-bit ID for the OMCI
// Managed entity Ethernet frame performance monitoring history data downstream
const EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID = ClassID(321) // 0x0141

var ethernetframeperformancemonitoringhistorydatadownstreamBME *ManagedEntityDefinition

// EthernetFramePerformanceMonitoringHistoryDataDownstream (Class ID: #321 / 0x0141)
//	This ME is identical to the Ethernet frame PM history data upstream ME, with the exception that
//	it monitors downstream traffic.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. Through an identical ID, this ME is
//			implicitly linked to an instance of a MAC bridge port configuration data. (R, setbycreate)
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
//		Drop Events
//			The total number of events in which packets were dropped due to a lack of resources. This is not
//			necessarily the number of packets dropped; it is the number of times this event was detected.
//			(R) (mandatory) (4-bytes)
//
//		Octets
//			The total number of upstream octets received, including those in bad packets, excluding framing
//			bits, but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets
//			The total number of upstream packets received, including bad packets, broadcast packets and
//			multicast packets. (R) (mandatory) (4-bytes)
//
//		Broadcast Packets
//			The total number of upstream good packets received that were directed to the broadcast address.
//			This does not include multicast packets. (R) (mandatory) (4-bytes)
//
//		Multicast Packets
//			The total number of upstream good packets received that were directed to a multicast address.
//			This does not include broadcast packets. (R) (mandatory) (4-bytes)
//
//		Crc Errored Packets
//			The total number of upstream packets received that had a length (excluding framing bits, but
//			including FCS octets) of between 64-octets and 1518-octets, inclusive, but had either a bad FCS
//			with an integral number of octets (FCS error) or a bad FCS with a non-integral number of octets
//			(alignment error). (R) (mandatory) (4-bytes)
//
//		Undersize Packets
//			The total number of upstream packets received that were less than 64-octets long, but were
//			otherwise well formed (excluding framing bits, but including FCS). (R) (mandatory) (4-bytes)
//
//		Oversize Packets
//			NOTE 2 - If 2-000-byte Ethernet frames are supported, counts in this performance parameter are
//			not necessarily errors.
//
//			The total number of upstream packets received that were longer than 1518-octets (excluding
//			framing bits, but including FCS) and were otherwise well formed. (R) (mandatory) (4-bytes)
//
//		Packets 64 Octets
//			The total number of upstream received packets (including bad packets) that were 64-octets long,
//			excluding framing bits but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets 65 To 127 Octets
//			The total number of upstream received packets (including bad packets) that were 65..127 octets
//			long, excluding framing bits but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets 128 To 255 Octets
//			The total number of upstream packets (including bad packets) received that were 128..255 octets
//			long, excluding framing bits but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets 256 To 511 Octets
//			The total number of upstream packets (including bad packets) received that were 256..511 octets
//			long, excluding framing bits but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets 512 To 1023 Octets
//			The total number of upstream packets (including bad packets) received that were 512..1-023
//			octets long, excluding framing bits but including FCS. (R) (mandatory) (4-bytes)
//
//		Packets 1024 To 1518 Octets
//			The total number of upstream packets (including bad packets) received that were 1024..1518
//			octets long, excluding framing bits, but including FCS. (R) (mandatory) (4-bytes)
//
type EthernetFramePerformanceMonitoringHistoryDataDownstream struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const EthernetFramePerformanceMonitoringHistoryDataDownstream_IntervalEndTime = "IntervalEndTime"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_ThresholdData12Id = "ThresholdData12Id"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_DropEvents = "DropEvents"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Octets = "Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets = "Packets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_BroadcastPackets = "BroadcastPackets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_MulticastPackets = "MulticastPackets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_CrcErroredPackets = "CrcErroredPackets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_UndersizePackets = "UndersizePackets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_OversizePackets = "OversizePackets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets64Octets = "Packets64Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets65To127Octets = "Packets65To127Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets128To255Octets = "Packets128To255Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets256To511Octets = "Packets256To511Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets512To1023Octets = "Packets512To1023Octets"
const EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets1024To1518Octets = "Packets1024To1518Octets"

func init() {
	ethernetframeperformancemonitoringhistorydatadownstreamBME = &ManagedEntityDefinition{
		Name:    "EthernetFramePerformanceMonitoringHistoryDataDownstream",
		ClassID: EthernetFramePerformanceMonitoringHistoryDataDownstreamClassID,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
			GetCurrentData,
		),
		AllowedAttributeMask: 0xffff,
		AttributeDefinitions: AttributeDefinitionMap{
			0:  Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1:  ByteField(EthernetFramePerformanceMonitoringHistoryDataDownstream_IntervalEndTime, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read), false, false, false, 1),
			2:  Uint16Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_ThresholdData12Id, PointerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_DropEvents, CounterAttributeType, 0x2000, 0, mapset.NewSetWith(Read), false, false, false, 3),
			4:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Octets, CounterAttributeType, 0x1000, 0, mapset.NewSetWith(Read), false, false, false, 4),
			5:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets, CounterAttributeType, 0x0800, 0, mapset.NewSetWith(Read), false, false, false, 5),
			6:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_BroadcastPackets, CounterAttributeType, 0x0400, 0, mapset.NewSetWith(Read), false, false, false, 6),
			7:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_MulticastPackets, CounterAttributeType, 0x0200, 0, mapset.NewSetWith(Read), false, false, false, 7),
			8:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_CrcErroredPackets, CounterAttributeType, 0x0100, 0, mapset.NewSetWith(Read), false, false, false, 8),
			9:  Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_UndersizePackets, CounterAttributeType, 0x0080, 0, mapset.NewSetWith(Read), false, false, false, 9),
			10: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_OversizePackets, CounterAttributeType, 0x0040, 0, mapset.NewSetWith(Read), false, false, false, 10),
			11: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets64Octets, CounterAttributeType, 0x0020, 0, mapset.NewSetWith(Read), false, false, false, 11),
			12: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets65To127Octets, CounterAttributeType, 0x0010, 0, mapset.NewSetWith(Read), false, false, false, 12),
			13: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets128To255Octets, CounterAttributeType, 0x0008, 0, mapset.NewSetWith(Read), false, false, false, 13),
			14: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets256To511Octets, CounterAttributeType, 0x0004, 0, mapset.NewSetWith(Read), false, false, false, 14),
			15: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets512To1023Octets, CounterAttributeType, 0x0002, 0, mapset.NewSetWith(Read), false, false, false, 15),
			16: Uint32Field(EthernetFramePerformanceMonitoringHistoryDataDownstream_Packets1024To1518Octets, CounterAttributeType, 0x0001, 0, mapset.NewSetWith(Read), false, false, false, 16),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
		Alarms: AlarmMap{
			0: "Drop events",
			1: "CRC errored packets",
			2: "Undersize packets",
			3: "Oversize packets",
		},
	}
}

// NewEthernetFramePerformanceMonitoringHistoryDataDownstream (class ID 321) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewEthernetFramePerformanceMonitoringHistoryDataDownstream(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*ethernetframeperformancemonitoringhistorydatadownstreamBME, params...)
}