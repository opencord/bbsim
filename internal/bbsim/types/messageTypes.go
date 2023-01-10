/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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

package types

import (
	"github.com/google/gopacket"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/omci-lib-go/v2"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"net"
)

type MessageType int

const (
	OltIndication     MessageType = 0
	NniIndication     MessageType = 1
	PonIndication     MessageType = 2
	OnuDiscIndication MessageType = 3
	OnuIndication     MessageType = 4
	OMCI              MessageType = 5
	FlowAdd           MessageType = 6
	FlowRemoved       MessageType = 18
	StartEAPOL        MessageType = 7
	StartDHCP         MessageType = 8
	OnuPacketOut      MessageType = 9

	// BBR messages
	OmciIndication MessageType = 10 // this are OMCI messages going from the OLT to VOLTHA
	SendEapolFlow  MessageType = 11
	SendDhcpFlow   MessageType = 12
	OnuPacketIn    MessageType = 13

	//IGMP
	IGMPMembershipReportV2 MessageType = 14 // Version 2 Membership Report (JOIN)
	IGMPLeaveGroup         MessageType = 15 // Leave Group

	AlarmIndication        MessageType = 16 // message data is an openolt.AlarmIndication
	IGMPMembershipReportV3 MessageType = 17 // Version 3 Membership Report
	UniStatusAlarm         MessageType = 19
)

func (m MessageType) String() string {
	names := [...]string{
		"OltIndication",
		"NniIndication",
		"PonIndication",
		"OnuDiscIndication",
		"OnuIndication",
		"OMCI",
		"FlowAdd",
		"StartEAPOL",
		"StartDHCP",
		"OnuPacketOut",
		"OmciIndication",
		"SendEapolFlow",
		"SendDhcpFlow",
		"OnuPacketIn",
		"IGMPMembershipReportV2",
		"IGMPLeaveGroup",
		"AlarmIndication",
		"IGMPMembershipReportV3",
		"FlowRemoved",
		"UniStatusAlarm",
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
}

type OnuIndicationMessage struct {
	OperState OperState
	PonPortID uint32
	OnuID     uint32
	OnuSN     *openolt.SerialNumber
}

// these are OMCI messages going from VOLTHA to the OLT
// used by BBSim
type OmciMessage struct {
	OnuSN   *openolt.SerialNumber
	OnuID   uint32
	OmciMsg *omci.OMCI
	OmciPkt gopacket.Packet
}

type UniStatusAlarmMessage struct {
	OnuSN          *openolt.SerialNumber
	OnuID          uint32
	AdminState     uint8
	EntityID       uint16
	RaiseOMCIAlarm bool
}

// these are OMCI messages going from the OLT to VOLTHA
// these messages are received by BBR
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
	IntfId     uint32
	OnuId      uint32
	PortNo     uint32
	Packet     gopacket.Packet
	Type       packetHandlers.PacketType
	MacAddress net.HardwareAddr
	GemPortId  uint32 // this is used by BBR
}

type IgmpMessage struct {
	GroupAddress string
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
