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

// MacBridgePortFilterPreAssignTableClassID is the 16-bit ID for the OMCI
// Managed entity MAC bridge port filter pre-assign table
const MacBridgePortFilterPreAssignTableClassID = ClassID(79) // 0x004f

var macbridgeportfilterpreassigntableBME *ManagedEntityDefinition

// MacBridgePortFilterPreAssignTable (Class ID: #79 / 0x004f)
//	This ME provides an alternate approach to DA filtering from that supported through the MAC
//	bridge port filter table data ME. This alternate approach is useful when all groups of addresses
//	are stored beforehand in the ONU, and the MAC bridge port filter pre-assign table ME designates
//	which groups are valid or invalid for filtering. On a circuit pack in which all groups of
//	addresses are pre-assigned and stored locally, the ONU creates or deletes an instance of this ME
//	automatically upon creation or deletion of a MAC bridge port configuration data ME.
//
//	Relationships
//		An instance of this ME is associated with an instance of a MAC bridge port configuration data
//		ME.
//
//	Attributes
//		Managed Entity Id
//			The following 10 attributes have similar definitions. Each permits the OLT to specify whether
//			MAC DAs or Ethertypes of the named type are forwarded (0) or filtered (1). In each case, the
//			initial value of the attribute is 0.
//
//			This attribute uniquely identifies each instance of this ME. Through an identical ID, this ME is
//			implicitly linked to an instance of the MAC bridge port configuration data ME. (R) (mandatory)
//			(2-bytes)
//
//		Ipv4 Multicast Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Ipv6 Multicast Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Ipv4 Broadcast Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Rarp Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Ipx Filtering
//				(R,-W) (mandatory) (1-byte)
//
//		Netbeui Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Appletalk Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Bridge Management Information Filtering
//			(R,-W) (mandatory) (1-byte)
//
//			Note that some destination MAC addresses should never be forwarded, considering the following
//			rules of [IEEE 802.1D].
//
//			1	Addresses from 01.80.C2.00.00.00 to 01.80.C2.00.00.0F are reserved.
//
//			2	Addresses from 01.80.C2.00.00.20 to 01.80.C2.00.00.2F are used for generic attribute
//			registration protocol (GARP) applications.
//
//		Arp Filtering
//			(R,-W) (mandatory) (1-byte)
//
//		Point_To_Point Protocol Over Ethernet Pppoe Broadcast Filtering
//			Point-to-point protocol over Ethernet (PPPoE) broadcast filtering:	(R,-W) (mandatory) (1-byte)
//
type MacBridgePortFilterPreAssignTable struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const MacBridgePortFilterPreAssignTable_Ipv4MulticastFiltering = "Ipv4MulticastFiltering"
const MacBridgePortFilterPreAssignTable_Ipv6MulticastFiltering = "Ipv6MulticastFiltering"
const MacBridgePortFilterPreAssignTable_Ipv4BroadcastFiltering = "Ipv4BroadcastFiltering"
const MacBridgePortFilterPreAssignTable_RarpFiltering = "RarpFiltering"
const MacBridgePortFilterPreAssignTable_IpxFiltering = "IpxFiltering"
const MacBridgePortFilterPreAssignTable_NetbeuiFiltering = "NetbeuiFiltering"
const MacBridgePortFilterPreAssignTable_AppletalkFiltering = "AppletalkFiltering"
const MacBridgePortFilterPreAssignTable_BridgeManagementInformationFiltering = "BridgeManagementInformationFiltering"
const MacBridgePortFilterPreAssignTable_ArpFiltering = "ArpFiltering"
const MacBridgePortFilterPreAssignTable_PointToPointProtocolOverEthernetPppoeBroadcastFiltering = "PointToPointProtocolOverEthernetPppoeBroadcastFiltering"

func init() {
	macbridgeportfilterpreassigntableBME = &ManagedEntityDefinition{
		Name:    "MacBridgePortFilterPreAssignTable",
		ClassID: MacBridgePortFilterPreAssignTableClassID,
		MessageTypes: mapset.NewSetWith(
			Get,
			Set,
		),
		AllowedAttributeMask: 0xffc0,
		AttributeDefinitions: AttributeDefinitionMap{
			0:  Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read), false, false, false, 0),
			1:  ByteField(MacBridgePortFilterPreAssignTable_Ipv4MulticastFiltering, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read, Write), false, false, false, 1),
			2:  ByteField(MacBridgePortFilterPreAssignTable_Ipv6MulticastFiltering, UnsignedIntegerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, Write), false, false, false, 2),
			3:  ByteField(MacBridgePortFilterPreAssignTable_Ipv4BroadcastFiltering, UnsignedIntegerAttributeType, 0x2000, 0, mapset.NewSetWith(Read, Write), false, false, false, 3),
			4:  ByteField(MacBridgePortFilterPreAssignTable_RarpFiltering, UnsignedIntegerAttributeType, 0x1000, 0, mapset.NewSetWith(Read, Write), false, false, false, 4),
			5:  ByteField(MacBridgePortFilterPreAssignTable_IpxFiltering, UnsignedIntegerAttributeType, 0x0800, 0, mapset.NewSetWith(Read, Write), false, false, false, 5),
			6:  ByteField(MacBridgePortFilterPreAssignTable_NetbeuiFiltering, UnsignedIntegerAttributeType, 0x0400, 0, mapset.NewSetWith(Read, Write), false, false, false, 6),
			7:  ByteField(MacBridgePortFilterPreAssignTable_AppletalkFiltering, UnsignedIntegerAttributeType, 0x0200, 0, mapset.NewSetWith(Read, Write), false, false, false, 7),
			8:  ByteField(MacBridgePortFilterPreAssignTable_BridgeManagementInformationFiltering, UnsignedIntegerAttributeType, 0x0100, 0, mapset.NewSetWith(Read, Write), false, false, false, 8),
			9:  ByteField(MacBridgePortFilterPreAssignTable_ArpFiltering, UnsignedIntegerAttributeType, 0x0080, 0, mapset.NewSetWith(Read, Write), false, false, false, 9),
			10: ByteField(MacBridgePortFilterPreAssignTable_PointToPointProtocolOverEthernetPppoeBroadcastFiltering, UnsignedIntegerAttributeType, 0x0040, 0, mapset.NewSetWith(Read, Write), false, false, false, 10),
		},
		Access:  CreatedByOnu,
		Support: UnknownSupport,
	}
}

// NewMacBridgePortFilterPreAssignTable (class ID 79) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewMacBridgePortFilterPreAssignTable(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*macbridgeportfilterpreassigntableBME, params...)
}
