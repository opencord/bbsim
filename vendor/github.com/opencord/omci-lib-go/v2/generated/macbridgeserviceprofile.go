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

// MacBridgeServiceProfileClassID is the 16-bit ID for the OMCI
// Managed entity MAC bridge service profile
const MacBridgeServiceProfileClassID = ClassID(45) // 0x002d

var macbridgeserviceprofileBME *ManagedEntityDefinition

// MacBridgeServiceProfile (Class ID: #45 / 0x002d)
//	This ME models a MAC bridge in its entirety; any number of ports may be associated with the
//	bridge through pointers to the MAC bridge service profile ME. Instances of this ME are created
//	and deleted by the OLT.
//
//	Relationships
//		Bridge ports are modelled by MAC bridge port configuration data MEs, any number of which can
//		point to a MAC bridge service profile. The real-time status of the bridge is available from an
//		implicitly linked MAC bridge configuration data ME.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. The first byte is the slot ID. In
//			an integrated ONU, this value is 0. The second byte is the bridge group ID. (R, setbycreate)
//			(mandatory) (2-bytes)
//
//		Spanning Tree Ind
//			The Boolean value true specifies that a spanning tree algorithm is enabled. The value false
//			disables (rapid) spanning tree. (R,-W, setbycreate) (mandatory) (1-byte)
//
//		Learning Ind
//			The Boolean value true specifies that bridge learning functions are enabled. The value false
//			disables bridge learning. (R,-W, setbycreate) (mandatory) (1-byte)
//
//		Port Bridging Ind
//			The Boolean value true specifies that bridging between UNI ports is enabled. The value false
//			disables local bridging. (R,-W, setbycreate) (mandatory) (1-byte)
//
//		Priority
//			This attribute specifies the bridge priority in the range 0..65535. The value of this attribute
//			is copied to the bridge priority attribute of the associated MAC bridge configuration data ME.
//			(R,-W, setbycreate) (mandatory) (2-bytes)
//
//		Max Age
//			This attribute specifies the maximum age (in 256ths of a second) of received protocol
//			information before its entry in the spanning tree listing is discarded. The range is 0x0600 to
//			0x2800 (6..40-s) in accordance with [IEEE-802.1D]. (R,-W, setbycreate) (mandatory) (2-bytes)
//
//		Hello Time
//			This attribute specifies how often (in 256ths of a second) the bridge advertises its presence
//			via hello packets, while acting as a root or attempting to become a root. The range is 0x0100 to
//			0x0A00 (1..10-s). (R,-W, setbycreate) (mandatory) (2-bytes)
//
//			NOTE - [IEEE 802.1D] specifies the compatibility range for hello time to be 1..2-s.
//
//		Forward Delay
//			This attribute specifies the forwarding delay (in 256ths of a second) when the bridge acts as
//			the root. The range is 0x0400 to 0x1E00 (4..30-s) in accordance with [IEEE 802.1D]. (R,-W, set-
//			by-create) (mandatory) (2-bytes)
//
//		Unknown Mac Address Discard
//			The Boolean value true specifies that MAC frames with unknown DAs be discarded. The value false
//			specifies that such frames be forwarded to all allowed ports. (R,-W, setbycreate) (mandatory)
//			(1-byte)
//
//		Mac Learning Depth
//			This attribute specifies the maximum number of UNI MAC addresses to be learned by the bridge.
//			The default value 0 specifies that there is no administratively imposed limit. (R,-W,
//			setbycreate) (optional) (1-byte)
//
//		Dynamic Filtering Ageing Time
//			This attribute specifies the age of dynamic filtering entries in the bridge database, after
//			which unrefreshed entries are discarded. In accordance with clause 7.9.2 of [IEEE 802.1D] and
//			clause 8.8.3 of [IEEE 802.1Q], the range is 10..1 000 000-s, with a resolution of 1-s and a
//			default of 300-s. The value 0 specifies that the ONU uses its internal default. (R, W, set-by-
//			create) (optional) (4 bytes)
//
type MacBridgeServiceProfile struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const MacBridgeServiceProfile_SpanningTreeInd = "SpanningTreeInd"
const MacBridgeServiceProfile_LearningInd = "LearningInd"
const MacBridgeServiceProfile_PortBridgingInd = "PortBridgingInd"
const MacBridgeServiceProfile_Priority = "Priority"
const MacBridgeServiceProfile_MaxAge = "MaxAge"
const MacBridgeServiceProfile_HelloTime = "HelloTime"
const MacBridgeServiceProfile_ForwardDelay = "ForwardDelay"
const MacBridgeServiceProfile_UnknownMacAddressDiscard = "UnknownMacAddressDiscard"
const MacBridgeServiceProfile_MacLearningDepth = "MacLearningDepth"
const MacBridgeServiceProfile_DynamicFilteringAgeingTime = "DynamicFilteringAgeingTime"

func init() {
	macbridgeserviceprofileBME = &ManagedEntityDefinition{
		Name:    "MacBridgeServiceProfile",
		ClassID: MacBridgeServiceProfileClassID,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
		),
		AllowedAttributeMask: 0xffc0,
		AttributeDefinitions: AttributeDefinitionMap{
			0:  Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1:  ByteField(MacBridgeServiceProfile_SpanningTreeInd, EnumerationAttributeType, 0x8000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 1),
			2:  ByteField(MacBridgeServiceProfile_LearningInd, EnumerationAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3:  ByteField(MacBridgeServiceProfile_PortBridgingInd, EnumerationAttributeType, 0x2000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 3),
			4:  Uint16Field(MacBridgeServiceProfile_Priority, UnsignedIntegerAttributeType, 0x1000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 4),
			5:  Uint16Field(MacBridgeServiceProfile_MaxAge, UnsignedIntegerAttributeType, 0x0800, 1536, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 5),
			6:  Uint16Field(MacBridgeServiceProfile_HelloTime, UnsignedIntegerAttributeType, 0x0400, 256, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 6),
			7:  Uint16Field(MacBridgeServiceProfile_ForwardDelay, UnsignedIntegerAttributeType, 0x0200, 1024, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 7),
			8:  ByteField(MacBridgeServiceProfile_UnknownMacAddressDiscard, EnumerationAttributeType, 0x0100, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 8),
			9:  ByteField(MacBridgeServiceProfile_MacLearningDepth, UnsignedIntegerAttributeType, 0x0080, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, true, false, 9),
			10: Uint32Field(MacBridgeServiceProfile_DynamicFilteringAgeingTime, UnsignedIntegerAttributeType, 0x0040, 300, mapset.NewSetWith(Read, SetByCreate, Write), false, true, false, 10),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
	}
}

// NewMacBridgeServiceProfile (class ID 45) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewMacBridgeServiceProfile(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*macbridgeserviceprofileBME, params...)
}
