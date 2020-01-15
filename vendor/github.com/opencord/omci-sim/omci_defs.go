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
package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// OMCI Sim definitions

type ChMessageType int

const (
	GemPortAdded ChMessageType = 0
	UniLinkUp ChMessageType = 1
	UniLinkDown ChMessageType = 2
)

func (m ChMessageType) String() string {
	names := [...]string{
		"GemPortAdded",
		"UniLinkUp",
		"UniLinkDown",
	}
	return names[m]
}

type OmciChMessageData struct {
	IntfId 	uint32
	OnuId   uint32
}

type OmciChMessage struct {
	Type ChMessageType
	Data OmciChMessageData
	Packet []byte
}

//
// OMCI definitions
//

// OmciMsgType represents a OMCI message-type
type OmciMsgType byte

func (t OmciMsgType) PrettyPrint() string {
	switch t {
	case Create:
		return "Create"
	case Delete:
		return "Delete"
	case Set:
		return "Set"
	case Get:
		return "Get"
	case GetAllAlarms:
		return "GetAllAlarms"
	case GetAllAlarmsNext:
		return "GetAllAlarmsNext"
	case MibUpload:
		return "MibUpload"
	case MibUploadNext:
		return "MibUploadNext"
	case MibReset:
		return "MibReset"
	case AlarmNotification:
		return "AlarmNotification"
	case AttributeValueChange:
		return "AttributeValueChange"
	case Test:
		return "Test"
	case StartSoftwareDownload:
		return "StartSoftwareDownload"
	case DownloadSection:
		return "DownloadSection"
	case EndSoftwareDownload:
		return "EndSoftwareDownload"
	case ActivateSoftware:
		return "ActivateSoftware"
	case CommitSoftware:
		return "CommitSoftware"
	case SynchronizeTime:
		return "SynchronizeTime"
	case Reboot:
		return "Reboot"
	case GetNext:
		return "GetNext"
	case TestResult:
		return "TestResult"
	case GetCurrentData:
		return "GetCurrentData"
	case SetTable:
		return "SetTable"
	default:
		// FIXME
		// msg="Cant't convert state 68 to string"
		// msg="Cant't convert state 72 to string"
		// msg="Cant't convert state 73 to string"
		// msg="Cant't convert state 75 to string"
		// msg="Cant't convert state 76 to string"
		// msg="Cant't convert state 77 to string"
		// msg="Cant't convert state 78 to string"
		// msg="Cant't convert state 79 to string"
		// msg="Cant't convert state 88 to string"

		log.Tracef("Cant't convert OmciMsgType %v to string", t)
		return string(t)
	}
}

const (
	// Message Types
	_                                 = iota
	Create                OmciMsgType = 4
	Delete                OmciMsgType = 6
	Set                   OmciMsgType = 8
	Get                   OmciMsgType = 9
	GetAllAlarms          OmciMsgType = 11
	GetAllAlarmsNext      OmciMsgType = 12
	MibUpload             OmciMsgType = 13
	MibUploadNext         OmciMsgType = 14
	MibReset              OmciMsgType = 15
	AlarmNotification     OmciMsgType = 16
	AttributeValueChange  OmciMsgType = 17
	Test                  OmciMsgType = 18
	StartSoftwareDownload OmciMsgType = 19
	DownloadSection       OmciMsgType = 20
	EndSoftwareDownload   OmciMsgType = 21
	ActivateSoftware      OmciMsgType = 22
	CommitSoftware        OmciMsgType = 23
	SynchronizeTime       OmciMsgType = 24
	Reboot                OmciMsgType = 25
	GetNext               OmciMsgType = 26
	TestResult            OmciMsgType = 27
	GetCurrentData        OmciMsgType = 28
	SetTable              OmciMsgType = 29 // Defined in Extended Message Set Only
)


// OMCI Managed Entity Class
type OmciClass uint16

func (c OmciClass) PrettyPrint() string {
	switch c {
	case EthernetPMHistoryData:
		return "EthernetPMHistoryData"
	case ONUG:
		return "ONUG"
	case ANIG:
		return "ANIG"
	case GEMPortNetworkCTP:
		return "GEMPortNetworkCTP"
	default:
		log.Tracef("Cant't convert OmciClass %v to string", c)
		return string(c)
	}
}

const (
	// Managed Entity Class values
	EthernetPMHistoryData OmciClass = 24
	ONUG                  OmciClass = 256
	ANIG                  OmciClass = 263
	GEMPortNetworkCTP     OmciClass = 268
)

// OMCI Message Identifier
type OmciMessageIdentifier struct {
	Class    OmciClass
	Instance uint16
}

type OmciContent [32]byte

type OmciMessage struct {
	TransactionId uint16
	MessageType   OmciMsgType
	DeviceId      uint8
	MessageId     OmciMessageIdentifier
	Content       OmciContent
}

func ParsePkt(pkt []byte) (uint16, uint8, OmciMsgType, OmciClass, uint16, OmciContent, error) {
	var m OmciMessage

	r := bytes.NewReader(pkt)

	if err := binary.Read(r, binary.BigEndian, &m); err != nil {
		log.WithFields(log.Fields{
			"Packet": pkt,
			"omciMsg": fmt.Sprintf("%x", pkt),
		}).Errorf("Failed to read packet: %s", err)
		return 0, 0, 0, 0, 0, OmciContent{}, errors.New("Failed to read packet")
	}
	/*    Message Type = Set
	      0... .... = Destination Bit: 0x0
	      .1.. .... = Acknowledge Request: 0x1
	      ..0. .... = Acknowledgement: 0x0
	      ...0 1000 = Message Type: Set (8)
	*/

	log.WithFields(log.Fields{
		"TransactionId": m.TransactionId,
		"MessageType": m.MessageType.PrettyPrint(),
		"MeClass": m.MessageId.Class,
		"MeInstance": m.MessageId.Instance,
		"Conent": m.Content,
		"Packet": pkt,
	}).Tracef("Parsing OMCI Packet")

	return m.TransactionId, m.DeviceId, m.MessageType & 0x1F, m.MessageId.Class, m.MessageId.Instance, m.Content, nil
}
