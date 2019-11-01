/*
 * Copyright 2018-present Open Networking Foundation

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

package devices

import (
	"github.com/google/gopacket"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/voltha-protos/go/openolt"
)

type MessageType int

const (
	OltIndication       MessageType = 0
	NniIndication       MessageType = 1
	PonIndication       MessageType = 2
	OnuDiscIndication   MessageType = 3
	OnuIndication       MessageType = 4
	OMCI                MessageType = 5
	FlowUpdate          MessageType = 6
	StartEAPOL          MessageType = 7
	StartDHCP           MessageType = 8
	OnuPacketOut        MessageType = 9
	DyingGaspIndication MessageType = 10

	// BBR messages
	OmciIndication MessageType = 11 // this are OMCI messages going from the OLT to VOLTHA
	SendEapolFlow  MessageType = 12
	SendDhcpFlow   MessageType = 13
	OnuPacketIn    MessageType = 14

	//IGMP
	IGMPMembershipReportV2 MessageType = 15 // Version 2 Membership Report (JOIN)
	IGMPLeaveGroup         MessageType = 16 // Leave Group
)

func (m MessageType) String() string {
	names := [...]string{
		"OltIndication",
		"NniIndication",
		"PonIndication",
		"OnuDiscIndication",
		"OnuIndication",
		"OMCI",
		"FlowUpdate",
		"StartEAPOL",
		"StartDHCP",
		"OnuPacketOut",
		"DyingGaspIndication",
		"OmciIndication",
		"SendEapolFlow",
		"SendDhcpFlow",
		"OnuPacketIn",
		"IGMPMembershipReportV2",
		"IGMPLeaveGroup",
	}
	return names[m]
}

type Message struct {
	Type MessageType
	Data interface{}
}

type OltIndicationMessage struct {
	OperState OperState
}

type NniIndicationMessage struct {
	OperState OperState
	NniPortID uint32
}

type PonIndicationMessage struct {
	OperState OperState
	PonPortID uint32
}

type OnuDiscIndicationMessage struct {
	OperState OperState
	Onu       *Onu
}

type OnuIndicationMessage struct {
	OperState OperState
	PonPortID uint32
	OnuID     uint32
	OnuSN     *openolt.SerialNumber
}

type OmciMessage struct {
	OnuSN   *openolt.SerialNumber
	OnuID   uint32
	omciMsg *openolt.OmciMsg
}

type OmciIndicationMessage struct {
	OnuSN   *openolt.SerialNumber
	OnuID   uint32
	OmciInd *openolt.OmciIndication
}

type OnuFlowUpdateMessage struct {
	PonPortID uint32
	OnuID     uint32
	Flow      *openolt.Flow
}

type PacketMessage struct {
	PonPortID uint32
	OnuID     uint32
}

type OnuPacketMessage struct {
	IntfId uint32
	OnuId  uint32
	Packet gopacket.Packet
	Type   packetHandlers.PacketType
}

type DyingGaspIndicationMessage struct {
	PonPortID uint32
	OnuID     uint32
	Status    string
}

type OperState int

const (
	UP   OperState = iota
	DOWN           // The device has been discovered, but not yet activated
)

func (m OperState) String() string {
	names := [...]string{
		"up",
		"down",
	}
	return names[m]
}
