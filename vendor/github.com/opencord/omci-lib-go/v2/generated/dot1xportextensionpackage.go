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

// Dot1XPortExtensionPackageClassID is the 16-bit ID for the OMCI
// Managed entity Dot1X port extension package
const Dot1XPortExtensionPackageClassID = ClassID(290) // 0x0122

var dot1xportextensionpackageBME *ManagedEntityDefinition

// Dot1XPortExtensionPackage (Class ID: #290 / 0x0122)
//	An instance of this ME represents a set of attributes that control a port's IEEE 802.1X
//	operation. It is created and deleted autonomously by the ONU upon the creation or deletion of a
//	PPTP that supports [IEEE 802.1X] authentication of customer premises equipment (CPE).
//
//	Relationships
//		An instance of this ME is associated with a PPTP that performs IEEE 802.1X authentication of CPE
//		(e.g., Ethernet or DSL).
//
//	Attributes
//		Managed Entity Id
//			This attribute provides a unique number for each instance of this ME. Its value is the same as
//			that of its associated PPTP (i.e., slot and port number). (R) (mandatory) (2-bytes)
//
//		Dot1X Enable
//			If true, this Boolean attribute forces the associated port to authenticate via [IEEE 802.1X] as
//			a precondition of normal service. The default value false does not impose IEEE 802.1X
//			authentication on the associated port. (R,-W) (mandatory) (1-byte)
//
//		Action Register
//			This attribute defines a set of actions that can be performed on the associated port. The act of
//			writing to the register causes the specified action.
//
//			1	Force re-authentication - this opcode initiates an IEEE-802.1X reauthentication conversation
//			with the associated port. The port remains in its current authorization state until the
//			conversation concludes.
//
//			2	Force unauthenticated - this opcode initiates an IEEE-802.1X authentication conversation whose
//			outcome is predestined to fail, thereby disabling normal Ethernet service on the port. The
//			port's provisioning is not changed, such that upon re-initialization, a new IEEE-802.1X
//			conversation may restore service without prejudice.
//
//			3	Force authenticated - this opcode initiates an IEEE-802.1X authentication conversation whose
//			outcome is predestined to succeed, thereby unconditionally enabling normal Ethernet service on
//			the port. The port's provisioning is not changed, such that upon re-initialization, a new
//			IEEE-802.1X conversation is required.
//
//			(W) (mandatory) (1-byte)
//
//		Authenticator Pae State
//			4	Authenticated
//
//			5	Aborting
//
//			6	Held
//
//			7	Force auth
//
//			8	Force unauth
//
//			9	Restart
//
//			(R) (optional) (1-byte)
//
//			This attribute returns the value of the port's PAE state. States are further described in [IEEE
//			802.1X]. Values are coded as follows.
//
//			0	Initialize
//
//			1	Disconnected
//
//			2	Connecting
//
//			3	Authenticating
//
//		Backend Authentication State
//			This attribute returns the value of the port's back-end authentication state. States are further
//			described in [IEEE 802.1X]. Values are coded as follows.
//
//			0	Request
//
//			1	Response
//
//			2	Success
//
//			3	Fail
//
//			4	Timeout
//
//			5	Idle
//
//			6	Initialize
//
//			7	Ignore
//
//			(R) (optional) (1-byte)
//
//		Admin Controlled Directions
//			This attribute controls the directionality of the port's authentication requirement. The default
//			value 0 indicates that control is imposed in both directions. The value 1 indicates that control
//			is imposed only on traffic from the subscriber towards the network. (R,-W) (optional) (1-byte)
//
//		Operational Controlled Directions
//			This attribute indicates the directionality of the port's current authentication state. The
//			value 0 indicates that control is imposed in both directions. The value 1 indicates that control
//			is imposed only on traffic from the subscriber towards the network. (R) (optional) (1-byte)
//
//		Authenticator Controlled Port Status
//			This attribute indicates whether the controlled port is currently authorized (1) or unauthorized
//			(2). (R) (optional) (1-byte)
//
//		Quiet Period
//			This attribute specifies the interval between EAP request/identity invitations sent to the peer.
//			Other events such as carrier present or EAPOL start frames from the peer may trigger an EAP
//			request/identity frame from the ONU at any time; this attribute controls the ONU's periodic
//			behaviour in the absence of these other inputs. It is expressed in seconds. (R,-W) (optional)
//			(2-bytes)
//
//		Server Timeout Period
//			This attribute specifies the time the ONU will wait for a response from the radius server before
//			timing out. Within this maximum interval, the ONU may initiate several retransmissions with
//			exponentially increasing delay. Upon timeout, the ONU may try another radius server if there is
//			one, or invoke the fallback policy, if no alternate radius servers are available. Server timeout
//			is expressed in seconds, with a default value of 30 and a maximum value of 65535. (R,-W)
//			(optional) (2-bytes)
//
//		Re_Authentication Period
//			Re-authentication period: This attribute records the re-authentication interval specified by the
//			radius authentication server. It is expressed in seconds. The attribute is only meaningful after
//			a port has been authenticated. (R) (optional) (2-bytes)
//
//		Re_Authentication Enabled
//			Re-authentication enabled: This Boolean attribute records whether the radius authentication
//			server has enabled re-authentication on this service (true) or not (false). The attribute is
//			only meaningful after a port has been authenticated. (R) (optional) (1-byte)
//
//		Key Transmission Enabled
//			This Boolean attribute indicates whether key transmission is enabled (true) or not (false). This
//			feature is not required; the parameter is listed here for completeness vis-`a-vis [IEEE 802.1X].
//			(R,-W) (optional) (1-byte)
//
type Dot1XPortExtensionPackage struct {
	ManagedEntityDefinition
	Attributes AttributeValueMap
}

// Attribute name constants

const Dot1XPortExtensionPackage_Dot1XEnable = "Dot1XEnable"
const Dot1XPortExtensionPackage_ActionRegister = "ActionRegister"
const Dot1XPortExtensionPackage_AuthenticatorPaeState = "AuthenticatorPaeState"
const Dot1XPortExtensionPackage_BackendAuthenticationState = "BackendAuthenticationState"
const Dot1XPortExtensionPackage_AdminControlledDirections = "AdminControlledDirections"
const Dot1XPortExtensionPackage_OperationalControlledDirections = "OperationalControlledDirections"
const Dot1XPortExtensionPackage_AuthenticatorControlledPortStatus = "AuthenticatorControlledPortStatus"
const Dot1XPortExtensionPackage_QuietPeriod = "QuietPeriod"
const Dot1XPortExtensionPackage_ServerTimeoutPeriod = "ServerTimeoutPeriod"
const Dot1XPortExtensionPackage_ReAuthenticationPeriod = "ReAuthenticationPeriod"
const Dot1XPortExtensionPackage_ReAuthenticationEnabled = "ReAuthenticationEnabled"
const Dot1XPortExtensionPackage_KeyTransmissionEnabled = "KeyTransmissionEnabled"

func init() {
	dot1xportextensionpackageBME = &ManagedEntityDefinition{
		Name:    "Dot1XPortExtensionPackage",
		ClassID: Dot1XPortExtensionPackageClassID,
		MessageTypes: mapset.NewSetWith(
			Get,
			Set,
		),
		AllowedAttributeMask: 0xfff0,
		AttributeDefinitions: AttributeDefinitionMap{
			0:  Uint16Field(ManagedEntityID, PointerAttributeType, 0x0000, 0, mapset.NewSetWith(Read), false, false, false, 0),
			1:  ByteField(Dot1XPortExtensionPackage_Dot1XEnable, UnsignedIntegerAttributeType, 0x8000, 0, mapset.NewSetWith(Read, Write), false, false, false, 1),
			2:  ByteField(Dot1XPortExtensionPackage_ActionRegister, UnsignedIntegerAttributeType, 0x4000, 0, mapset.NewSetWith(Write), false, false, false, 2),
			3:  ByteField(Dot1XPortExtensionPackage_AuthenticatorPaeState, UnsignedIntegerAttributeType, 0x2000, 0, mapset.NewSetWith(Read), false, true, false, 3),
			4:  ByteField(Dot1XPortExtensionPackage_BackendAuthenticationState, UnsignedIntegerAttributeType, 0x1000, 0, mapset.NewSetWith(Read), false, true, false, 4),
			5:  ByteField(Dot1XPortExtensionPackage_AdminControlledDirections, UnsignedIntegerAttributeType, 0x0800, 0, mapset.NewSetWith(Read, Write), false, true, false, 5),
			6:  ByteField(Dot1XPortExtensionPackage_OperationalControlledDirections, UnsignedIntegerAttributeType, 0x0400, 0, mapset.NewSetWith(Read), false, true, false, 6),
			7:  ByteField(Dot1XPortExtensionPackage_AuthenticatorControlledPortStatus, UnsignedIntegerAttributeType, 0x0200, 0, mapset.NewSetWith(Read), false, true, false, 7),
			8:  Uint16Field(Dot1XPortExtensionPackage_QuietPeriod, UnsignedIntegerAttributeType, 0x0100, 0, mapset.NewSetWith(Read, Write), false, true, false, 8),
			9:  Uint16Field(Dot1XPortExtensionPackage_ServerTimeoutPeriod, UnsignedIntegerAttributeType, 0x0080, 0, mapset.NewSetWith(Read, Write), false, true, false, 9),
			10: Uint16Field(Dot1XPortExtensionPackage_ReAuthenticationPeriod, UnsignedIntegerAttributeType, 0x0040, 0, mapset.NewSetWith(Read), false, true, false, 10),
			11: ByteField(Dot1XPortExtensionPackage_ReAuthenticationEnabled, UnsignedIntegerAttributeType, 0x0020, 0, mapset.NewSetWith(Read), false, true, false, 11),
			12: ByteField(Dot1XPortExtensionPackage_KeyTransmissionEnabled, UnsignedIntegerAttributeType, 0x0010, 0, mapset.NewSetWith(Read, Write), false, true, false, 12),
		},
		Access:  CreatedByOnu,
		Support: UnknownSupport,
		Alarms: AlarmMap{
			0: "dot1x local authentication - allowed",
			1: "dot1x local authentication - denied",
		},
	}
}

// NewDot1XPortExtensionPackage (class ID 290) creates the basic
// Managed Entity definition that is used to validate an ME of this type that
// is received from or transmitted to the OMCC.
func NewDot1XPortExtensionPackage(params ...ParamData) (*ManagedEntity, OmciErrors) {
	return NewManagedEntity(*dot1xportextensionpackageBME, params...)
}
