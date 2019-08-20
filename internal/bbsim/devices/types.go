package devices

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/opencord/voltha-protos/go/openolt"
	"github.com/looplab/fsm"
)

// Devices
type Onu struct {
	ID uint32
	PonPortID uint32
	PonPort PonPort
	InternalState *fsm.FSM

	OperState *fsm.FSM
	SerialNumber *openolt.SerialNumber

	channel chan Message
}



type NniPort struct {
	// BBSIM Internals
	ID uint32

	// PON Attributes
	OperState *fsm.FSM
	Type string
}

type PonPort struct {
	// BBSIM Internals
	ID uint32
	NumOnu int
	Onus []Onu

	// PON Attributes
	OperState *fsm.FSM
	Type string

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
	ID int
	NumNni int
	NumPon int
	NumOnuPerPon int
	InternalState *fsm.FSM
	channel chan Message

	Pons []PonPort
	Nnis []NniPort

	// OLT Attributes
	OperState *fsm.FSM
}

// BBSim Internals
type MessageType int

const (
	OltIndication     MessageType = 0
	NniIndication     MessageType = 1
	PonIndication     MessageType = 2
	OnuDiscIndication MessageType = 3
	OnuIndication     MessageType = 4
	OMCI              MessageType = 5
)

func (m MessageType) String() string {
	names := [...]string{
		"OltIndication",
		"NniIndication",
		"PonIndication",
		"OnuDiscIndication",
		"OnuIndication",
		"OMCI",
	}
	return names[m]
}

type Message struct {
	Type      MessageType
	Data 	  interface{}
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
	OnuSN     *openolt.SerialNumber
	OnuID 	  uint32
	omciMsg   *openolt.OmciMsg
}


type OperState int

const (
	UP OperState = iota
	DOWN // The device has been discovered, but not yet activated
)

func (m OperState) String() string {
	names := [...]string{
		"up",
		"down",
	}
	return names[m]
}