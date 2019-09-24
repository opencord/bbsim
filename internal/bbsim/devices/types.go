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
	"bytes"
	"errors"
	"fmt"
	"github.com/looplab/fsm"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/go/openolt"
	"net"
)

// TODO get rid of this file
// - move ONU and OLT struct in their respective file
// - create files for PonPorts and NniPorts
// - move messages in the `types` package

// Devices
type Onu struct {
	ID            uint32
	PonPortID     uint32
	PonPort       PonPort
	STag          int
	CTag          int
	HwAddress     net.HardwareAddr
	InternalState *fsm.FSM

	OperState    *fsm.FSM
	SerialNumber *openolt.SerialNumber

	Channel       chan Message        // this Channel is to track state changes and OMCI messages
	eapolPktOutCh chan *bbsim.ByteMsg // this Channel is for EAPOL Packet Outs (coming from the controller)
	dhcpPktOutCh  chan *bbsim.ByteMsg // this Channel is for DHCP Packet Outs (coming from the controller)
}

func (o Onu) Sn() string {
	return onuSnToString(o.SerialNumber)
}

type NniPort struct {
	// BBSIM Internals
	ID uint32

	// PON Attributes
	OperState *fsm.FSM
	Type      string
}

type PonPort struct {
	// BBSIM Internals
	ID     uint32
	NumOnu int
	Onus   []Onu
	Olt    OltDevice

	// PON Attributes
	OperState *fsm.FSM
	Type      string

	// NOTE do we need a state machine for the PON Ports?
}

func (p PonPort) getOnuBySn(sn *openolt.SerialNumber) (*Onu, error) {
	for _, onu := range p.Onus {
		if bytes.Equal(onu.SerialNumber.VendorSpecific, sn.VendorSpecific) {
			return &onu, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find Onu with serial number %d in PonPort %d", sn, p.ID))
}

func (p PonPort) getOnuById(id uint32) (*Onu, error) {
	for _, onu := range p.Onus {
		if onu.ID == id {
			return &onu, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find Onu with id %d in PonPort %d", id, p.ID))
}

type OltDevice struct {
	// BBSIM Internals
	ID              int
	SerialNumber    string
	NumNni          int
	NumPon          int
	NumOnuPerPon    int
	InternalState   *fsm.FSM
	channel         chan Message
	oltDoneChannel  *chan bool
	apiDoneChannel  *chan bool
	nniPktInChannel chan *bbsim.PacketMsg

	Pons []PonPort
	Nnis []NniPort

	// OLT Attributes
	OperState *fsm.FSM
}

// BBSim Internals

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
	DyingGaspIndication MessageType = 9
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
		"DyingGaspIndication",
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
	Onu       Onu
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

type OnuFlowUpdateMessage struct {
	PonPortID uint32
	OnuID     uint32
	Flow      *openolt.Flow
}

type PacketMessage struct {
	PonPortID uint32
	OnuID     uint32
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
