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

// SipAgentConfigData2ClassID is the 16-bit ID for the OMCI
// Managed entity SIP agent config data 2
const SipAgentConfigData2ClassID = ClassID(407) // 0x0197

var sipagentconfigdata2BME *ManagedEntityDefinition

// SipAgentConfigData2 (Class ID: #407 / 0x0197)
//	This ME supplements SIP agent config data ME. Instances of this ME are created and deleted by
//	the OLT.
//
//	Relationships
//		An instance of this ME is associated with a SIP agent config data.
//
//	Attributes
//		Managed Entity Id
//			This attribute uniquely identifies each instance of this ME. Through an identical ID, this ME is
//			implicitly linked to an instance of the corresponding SIP agent config data.  Note that this
//			entity is associated with the primary SIP agent config data (if SIP agent is involved in
//			protection). (R, set-by-create) (mandatory) (2-bytes)
//
//		In_Use_Options_Timer
//			In-Use-Options-Timer: This attribute defines the frequency that a SIP options packet is sent to
//			the SIP proxy in-use. When a SIP options packet is not responded to by the SIP proxy, it is
//			marked as unavailable. Otherwise, it is marked as available. Units are seconds. The default
//			value 0 specifies vendor-specific implementation. (R, W, set-by-create) (mandatory) (2-byte)
//
//		Alternate_Options_Timer
//			Alternate-Options-Timer: This attribute defines the frequency that a SIP options packet is sent
//			to the standby SIP proxy. When a SIP options packet is not responded to by the standby SIP
//			proxy, it is marked as unavailable. Otherwise, it is marked as available. Units are seconds. The
//			default value 0 specifies vendor-specific implementation. (R, W, set-by-create) (mandatory)
//			(2-byte)
//
//		Revertive
//			This Boolean attribute specifies whether the SIP UA is involved in revertive (true) or non-
//			revertive (false) switching. The default value is recommended to be false. (R, W, set-by-create)
//			(mandatory) (1 byte)
//
//		Current Proxy Server Resolved Address
//			This attribute contains the resolved IP address of the in-use SIP proxy. If the value is
//			0.0.x.y, where x and y are not both 0, then x.y is to be interpreted as a pointer to a large
//			string ME that represents an IPv6 address. Otherwise, the address is an IPv4 address (R)
//			(optional) (4-bytes)
//
//		Current Proxy Server Resolved Name
//			This attribute contains a pointer to the large string ME that contains the resolved name of the
//			SIP proxy in-use. (R) (optional) (2-bytes)
//
//		Alternate Proxy Server Resolved Address
//			This attribute contains the resolved IP address of the alternate SIP proxy. If the value is
//			0.0.x.y, where x and y are not both 0, then x.y is to be interpreted as a pointer to a large
//			string ME that represents an IPv6 address. Otherwise, the address is an IPv4 address (R)
//			(optional) (4-bytes)
//
//		Alternate Proxy Server Resolved Name
//			This attribute contains a pointer to the large string ME that contains the resolved name of the
//			alternate SIP proxy. (R) (optional) (2-bytes)
//
type SipAgentConfigData2 struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const SipAgentConfigData2_InUseOptionsTimer = "InUseOptionsTimer"
const SipAgentConfigData2_AlternateOptionsTimer = "AlternateOptionsTimer"
const SipAgentConfigData2_Revertive = "Revertive"
const SipAgentConfigData2_CurrentProxyServerResolvedAddress = "CurrentProxyServerResolvedAddress"
const SipAgentConfigData2_CurrentProxyServerResolvedName = "CurrentProxyServerResolvedName"
const SipAgentConfigData2_AlternateProxyServerResolvedAddress = "AlternateProxyServerResolvedAddress"
const SipAgentConfigData2_AlternateProxyServerResolvedName = "AlternateProxyServerResolvedName"

func init() {
	sipagentconfigdata2BME = &ManagedEntityDefinition{
		Name:    "SipAgentConfigData2",
		ClassID: SipAgentConfigData2ClassID,
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
			1: Uint16Field(SipAgentConfigData2_InUseOptionsTimer, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 1),
			2: Uint16Field(SipAgentConfigData2_AlternateOptionsTimer, UnsignedIntegerAttributeType, 0x4000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 2),
			3: ByteField(SipAgentConfigData2_Revertive, UnsignedIntegerAttributeType, 0x2000, 0, mapset.NewSetWith(Read, SetByCreate, Write), false, false, false, 3),
			4: Uint32Field(SipAgentConfigData2_CurrentProxyServerResolvedAddress, UnsignedIntegerAttributeType, 0x1000, 0, mapset.NewSetWith(Read), false, true, false, 4),
			5: Uint16Field(SipAgentConfigData2_CurrentProxyServerResolvedName, UnsignedIntegerAttributeType, 0x0800, 0, mapset.NewSetWith(Read), false, true, false, 5),
			6: Uint32Field(SipAgentConfigData2_AlternateProxyServerResolvedAddress, UnsignedIntegerAttributeType, 0x0400, 0, mapset.NewSetWith(Read), false, true, false, 6),
			7: Uint16Field(SipAgentConfigData2_AlternateProxyServerResolvedName, UnsignedIntegerAttributeType, 0x0200, 0, mapset.NewSetWith(Read), false, true, false, 7),
		},
		Access:  CreatedByOlt,
		Support: UnknownSupport,
	}
}

// NewSipAgentConfigData2 (class ID 407) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewSipAgentConfigData2(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*sipagentconfigdata2BME, params...)
}
