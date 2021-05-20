/*
 * Copyright (c) 2018 - present.  Boling Consulting Solutions (bcsw.net)
 * Copyright 2020-present Open Networking Foundation

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
/*
 * NOTE: This file was generated, manual edits will be overwritten!
 *
 * Generated by 'goCodeGenerator.py':
 *              https://github.com/cboling/OMCI-parser/README.md
 */

package generated

import "github.com/deckarep/golang-set"

// TContClassID is the 16-bit ID for the OMCI
// Managed entity T-CONT
const TContClassID ClassID = ClassID(262)

var tcontBME *ManagedEntityDefinition

// TCont (class ID #262)
//	An instance of the traffic container ME T-CONT represents a logical connection group associated
//	with a G-PON PLOAM layer alloc-ID. A T-CONT can accommodate GEM packets in priority queues or
//	traffic schedulers that exist in the GEM layer.
//
//	The ONU autonomously creates instances of this ME. The OLT can discover the number of TCONT
//	instances via the ANI-G ME. When the ONU's MIB is reset or created for the first time, all
//	supported T-CONTs are created. The OLT provisions alloc-IDs to the ONU via the PLOAM channel.
//	Via the OMCI, the OLT must then set the alloc-ID attributes in the T-CONTs that it wants to
//	activate for user traffic, to create the appropriate association with the allocation ID in the
//	PLOAM channel. There should be a one-to-one relationship between allocation IDs and T-CONT MEs;
//	the connection of multiple T-CONTs to a single allocation ID is undefined.
//
//	The allocation ID that matches the ONU-ID itself is defined to be the default alloc-ID. This
//	allocID is used to carry the OMCC. The default alloc-ID can also be used to carry user traffic,
//	and hence can be assigned to one of the T-CONT MEs. However, this OMCI relationship only
//	pertains to user traffic, and the OMCC relationship is unaffected. It can also be true that the
//	OMCC is not contained in any T-CONT ME construct; rather, that the OMCC remains outside of the
//	OMCI, and that the OMCI is not used to manage the OMCC in any way. Multiplexing of the OMCC and
//	user data in GPON systems is discussed in clause B.2.4.
//
//	Relationships
//		One or more instances of this ME are associated with an instance of a circuit pack that supports
//		a PON interface function, or with the ONU-G itself.
//
//	Attributes
//		Managed Entity Id
//			Managed entity ID: This attribute uniquely identifies each instance of this ME. This 2-byte
//			number indicates the physical capability that realizes the TCONT. It may be represented as
//			0xSSBB, where SS indicates the slot ID that contains this T-CONT (0 for the ONU as a whole), and
//			BB is the TCONT ID, numbered by the ONU itself. T-CONTs are numbered in ascending order, with
//			the range 0..255 in each slot. (R) (mandatory) (2-bytes)
//
//		Alloc_Id
//			Alloc-ID:	This attribute links the T-CONT with the alloc-ID assigned by the OLT in the
//			assign_alloc-ID PLOAM message. The respective TC layer specification should be referenced for
//			the legal values for that system. Prior to the setting of this attribute by the OLT, this
//			attribute has an unambiguously unusable initial value, namely the value 0x00FF or 0xFFFF for
//			ITU-T G.984 systems, and the value 0xFFFF for all other ITU-T GTC based PON systems. (R,-W)
//			(mandatory) (2-bytes)
//
//		Deprecated
//			Deprecated:	The ONU should set this attribute to the value 1, and the OLT should ignore it. (R)
//			(mandatory) (1-byte)
//
//		Policy
//			NOTE - This attribute is read-only, unless otherwise specified by the QoS configuration
//			flexibility attribute of the ONU2-G ME. If flexible configuration is not supported, the ONU
//			should reject an attempt to set it with a parameter error result-reason code.
//
type TCont struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

func init() {
	tcontBME = &ManagedEntityDefinition{
		Name:    "TCont",
		ClassID: 262,
		MessageTypes: mapset.NewSetWith(
			Get,
			Set,
		),
		AllowedAttributeMask: 0xe000,
		AttributeDefinitions: AttributeDefinitionMap{
			0: Uint16Field("ManagedEntityId", PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read), false, false, false, 0),
			1: Uint16Field("AllocId", UnsignedIntegerAttributeType, 0x8000, 65535, mapset.NewSetWith(Read, Write), false, false, false, 1),
			2: ByteField("Deprecated", UnsignedIntegerAttributeType, 0x4000, 1, mapset.NewSetWith(Read), false, false, true, 2),
			3: ByteField("Policy", EnumerationAttributeType, 0x2000, 0, mapset.NewSetWith(Read, Write), false, false, false, 3),
		},
		Access:  CreatedByOnu,
		Support: UnknownSupport,
	}
}

// NewTCont (class ID 262) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewTCont(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*tcontBME, params...)
}