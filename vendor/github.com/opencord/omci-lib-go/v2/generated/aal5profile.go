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

// Aal5ProfileClassID is the 16-bit ID for the OMCI
// Managed entity AAL5 profile
const Aal5ProfileClassID = ClassID(16) // 0x0010

var aal5profileBME *ManagedEntityDefinition

// Aal5Profile (Class ID: #16 / 0x0010)
//	This ME organizes data that describe the AAL type 5 processing functions of the ONU. It is used
//	with the IW VCC TP ME.
//
//	This ME is created and deleted by the OLT.
//
//	Relationships
//		An instance of this ME may be associated with zero or more instances of the IW VCC TP.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. (R, setbycreate) (mandatory)
//			(2-bytes)
//
//		Max Cpcs Pdu Size
//			This attribute specifies the maximum CPCS PDU size to be transmitted over the connection in both
//			upstream and downstream directions. (R,-W, setbycreate) (mandatory) (2-bytes)
//
//		Aal Mode
//			This attribute specifies the AAL mode as follows.
//
//			0	Message assured
//
//			1	Message unassured
//
//			2	Streaming assured
//
//			3	Streaming non assured
//
//			(R,-W, setbycreate) (mandatory) (1-byte)
//
//		Sscs Type
//			This attribute specifies the SSCS type for the AAL. Valid values are as follows.
//
//			0	Null
//
//			1	Data SSCS based on SSCOP, assured operation
//
//			2	Data SSCS based on SSCOP, non-assured operation
//
//			3	Frame relay SSCS
//
//			(R,-W, setbycreate) (mandatory) (1-byte)
//
type Aal5Profile struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const Aal5Profile_MaxCpcsPduSize = "MaxCpcsPduSize"
const Aal5Profile_AalMode = "AalMode"
const Aal5Profile_SscsType = "SscsType"

func init() {
	aal5profileBME = &ManagedEntityDefinition{
		Name:    "Aal5Profile",
		ClassID: Aal5ProfileClassID,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
		),
		AllowedAttributeMask: 0xe000,
		AttributeDefinitions: AttributeDefinitionMap{
			0: Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1: Uint16Field(Aal5Profile_MaxCpcsPduSize, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 1),
			2: ByteField(Aal5Profile_AalMode, UnsignedIntegerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3: ByteField(Aal5Profile_SscsType, UnsignedIntegerAttributeType, 0x2000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 3),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
	}
}

// NewAal5Profile (class ID 16) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewAal5Profile(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*aal5profileBME, params...)
}