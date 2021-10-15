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

// LargeStringClassID is the 16-bit ID for the OMCI
// Managed entity Large string
const LargeStringClassID = ClassID(157) // 0x009d

var largestringBME *ManagedEntityDefinition

// LargeString (Class ID: #157 / 0x009d)
//	The large string ME holds character strings longer than 25-bytes, up to 375-bytes. It is
//	maintained in up to 15 parts, each part containing 25-bytes. If the final part contains fewer
//	than 25-bytes, it is terminated by at least one null byte. For example:
//
//	Or
//
//	Instances of this ME are created and deleted by the OLT. Under some circumstances, they may also
//	be created by the ONU. To use this ME, the OLT or ONU instantiates the large string ME and then
//	points to the created ME from other ME instances. Systems that maintain the large string should
//	ensure that the large string ME is not deleted while it is still linked.
//
//	Relationships
//		An instance of this ME may be cited by any ME that requires a text string longer than 25-bytes.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. The value 0xFFFF is reserved. When
//			the large string is to be used as an IPv6 address, the value 0 is also reserved. The OLT should
//			create large string MEs starting at 1 (or 0), and numbering upwards. The ONU should create large
//			string MEs starting at 65534 (0xFFFE) and numbering downwards. (R,-setbycreate) (mandatory)
//			(2-bytes)
//
//		Number Of Parts
//			This attribute specifies the number of non-empty parts that form the large string. This
//			attribute defaults to 0 to indicate no large string content is defined.(R,-W) (mandatory)
//			(1-byte)
//
//			Fifteen additional attributes are defined in the following; they are identical. The large string
//			is simply divided into as many parts as necessary, starting at part 1. If the end of the string
//			does not lie at a part boundary, it is marked with a null byte.
//
//		Part 1
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 2
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 3
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 4
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 5
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 6
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 7
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 8
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 9
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 10
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 11
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 12
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 13
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 14
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
//		Part 15
//			Part 1, Part 2, Part 3, Part 4, Part 5, Part 6, Part 7, Part 8, Part 9,  Part 10, Part 11, Part
//			12, Part 13, Part 14, Part 15: (R,-W) (mandatory) (25-bytes * 15 attributes)
//
type LargeString struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

func init() {
	largestringBME = &ManagedEntityDefinition{
		Name:    "LargeString",
		ClassID: 157,
		MessageTypes: mapset.NewSetWith(
			Create,
			Delete,
			Get,
			Set,
		),
		AllowedAttributeMask: 0xffff,
		AttributeDefinitions: AttributeDefinitionMap{
			0:  Uint16Field("ManagedEntityId", PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read, SetByCreate), false, false, false, 0),
			1:  ByteField("NumberOfParts", UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read, Write), true, false, false, 1),
			2:  MultiByteField("Part1", OctetsAttributeType, 0x4000, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 2),
			3:  MultiByteField("Part2", OctetsAttributeType, 0x2000, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 3),
			4:  MultiByteField("Part3", OctetsAttributeType, 0x1000, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 4),
			5:  MultiByteField("Part4", OctetsAttributeType, 0x0800, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 5),
			6:  MultiByteField("Part5", OctetsAttributeType, 0x0400, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 6),
			7:  MultiByteField("Part6", OctetsAttributeType, 0x0200, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 7),
			8:  MultiByteField("Part7", OctetsAttributeType, 0x0100, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 8),
			9:  MultiByteField("Part8", OctetsAttributeType, 0x0080, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 9),
			10: MultiByteField("Part9", OctetsAttributeType, 0x0040, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 10),
			11: MultiByteField("Part10", OctetsAttributeType, 0x0020, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 11),
			12: MultiByteField("Part11", OctetsAttributeType, 0x0010, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 12),
			13: MultiByteField("Part12", OctetsAttributeType, 0x0008, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 13),
			14: MultiByteField("Part13", OctetsAttributeType, 0x0004, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 14),
			15: MultiByteField("Part14", OctetsAttributeType, 0x0002, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 15),
			16: MultiByteField("Part15", OctetsAttributeType, 0x0001, 25, toOctets("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="), mapset.NewSetWith(Read, Write), true, false, false, 16),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
	}
}

// NewLargeString (class ID 157) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewLargeString(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*largestringBME, params...)
}
