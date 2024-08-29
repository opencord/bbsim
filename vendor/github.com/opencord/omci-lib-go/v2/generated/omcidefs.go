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

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/google/gopacket"
)

// MsgType represents a OMCI message-type
type MsgType byte

// MsgType represents the status field in a OMCI Response frame
type Results byte

// AttributeAccess represents the access allowed to an Attribute.  Some MEs
// are instantiated by the ONU autonomously. Others are instantiated on
// explicit request of the OLT via a create command, and a few ME types may
// be instantiated in either way, depending on the ONU architecture or
// circumstances.
//
// Attributes of an ME that is auto-instantiated by the ONU can be read (R),
// write (W), or read, write (R, W). On the other hand, attributes of a ME
// that is instantiated by the OLT can be either (R), (W), (R, W),
// (R, set by create) or (R, W, set by create).
type AttributeAccess byte

// ClassID is a 16-bit value that uniquely defines a Managed Entity clas
// from the ITU-T G.988 specification.
type ClassID uint16

// AlarmMap is a mapping of alarm bit numbers to alarm names and can be
// used during decode of Alarm Notification messages
type AlarmMap map[uint8]string

func (cid ClassID) String() string {
	if entity, err := LoadManagedEntityDefinition(cid); err.StatusCode() == Success {
		return fmt.Sprintf("[%s] (%d/%#x)",
			entity.GetManagedEntityDefinition().GetName(), uint16(cid), uint16(cid))
	}
	return fmt.Sprintf("unknown ClassID: %d (%#x)", uint16(cid), uint16(cid))
}

const (
	// AK (Bit 6), indicates whether this message is an AK to an action request.
	// If a message is an AK, this bit is set to 1. If the message is not a
	// response to a command, this bit is set to 0. In messages sent by the OLT,
	// this bit is always 0.
	AK byte = 0x20

	// AR (Bit 7), acknowledge request, indicates whether the message requires an
	// AK. An AK is a response to an action request, not a link layer handshake.
	// If an AK is expected, this bit is set to 1. If no AK is expected, this bit
	// is 0. In messages sent by the ONU, this bit is always 0
	AR byte = 0x40

	// MsgTypeMask provides a mask to get the base message type
	MsgTypeMask = 0x1F
)

const (
	// Message Types
	Create                MsgType = 4
	Delete                MsgType = 6
	Set                   MsgType = 8
	Get                   MsgType = 9
	GetAllAlarms          MsgType = 11
	GetAllAlarmsNext      MsgType = 12
	MibUpload             MsgType = 13
	MibUploadNext         MsgType = 14
	MibReset              MsgType = 15
	AlarmNotification     MsgType = 16
	AttributeValueChange  MsgType = 17
	Test                  MsgType = 18
	StartSoftwareDownload MsgType = 19
	DownloadSection       MsgType = 20
	EndSoftwareDownload   MsgType = 21
	ActivateSoftware      MsgType = 22
	CommitSoftware        MsgType = 23
	SynchronizeTime       MsgType = 24
	Reboot                MsgType = 25
	GetNext               MsgType = 26
	TestResult            MsgType = 27
	GetCurrentData        MsgType = 28
	SetTable              MsgType = 29 // Defined in Extended Message Set Only
	ExtendedOffset        MsgType = 0x80
)

func (mt MsgType) String() string {
	switch mt {
	default:
		return "Unknown"
	case Create:
		return "Create"
	case Delete:
		return "Delete"
	case Set:
		return "Set"
	case Get:
		return "Get"
	case GetAllAlarms:
		return "Get All Alarms"
	case GetAllAlarmsNext:
		return "Get All Alarms Next"
	case MibUpload:
		return "MIB Upload"
	case MibUploadNext:
		return "MIB Upload Next"
	case MibReset:
		return "MIB Reset"
	case AlarmNotification:
		return "Alarm Notification"
	case AttributeValueChange:
		return "Attribute Value Change"
	case Test:
		return "Test"
	case StartSoftwareDownload:
		return "Start Software Download"
	case DownloadSection:
		return "Download Section"
	case EndSoftwareDownload:
		return "EndSoftware Download"
	case ActivateSoftware:
		return "Activate Software"
	case CommitSoftware:
		return "Commit Software"
	case SynchronizeTime:
		return "Synchronize Time"
	case Reboot:
		return "Reboot"
	case GetNext:
		return "Get Next"
	case TestResult:
		return "Test Result"
	case GetCurrentData:
		return "Get Current Data"
	case SetTable:
		return "Set Table"
	}
}

var allNotificationTypes = [...]MsgType{
	AlarmNotification,
	AttributeValueChange,
	TestResult,
}

// SupportsMsgType returns true if the managed entity supports the desired
// Message Type / action
func SupportsMsgType(entity IManagedEntityDefinition, msgType MsgType) bool {
	return entity.GetMessageTypes().Contains(msgType)
}

func IsAutonomousNotification(mt MsgType) bool {
	for _, m := range allNotificationTypes {
		if mt == m {
			return true
		}
	}
	return false
}

const (
	// Response status codes
	Success          Results = 0 // command processed successfully
	ProcessingError  Results = 1 // command processing error
	NotSupported     Results = 2 // command not supported
	ParameterError   Results = 3 // parameter error
	UnknownEntity    Results = 4 // unknown managed entity
	UnknownInstance  Results = 5 // unknown managed entity instance
	DeviceBusy       Results = 6 // device busy
	InstanceExists   Results = 7 // instance exists
	AttributeFailure Results = 9 // Attribute(s) failed or unknown
)

func (rc Results) String() string {
	switch rc {
	default:
		return "Unknown"
	case Success:
		return "Success"
	case ProcessingError:
		return "Processing Error"
	case NotSupported:
		return "Not Supported"
	case ParameterError:
		return "Parameter Error"
	case UnknownEntity:
		return "Unknown Entity"
	case UnknownInstance:
		return "Unknown Instance"
	case DeviceBusy:
		return "Device Busy"
	case InstanceExists:
		return "Instance Exists"
	case AttributeFailure:
		return "Attribute Failure"
	}
}

const (
	// Access allowed on a Managed Entity attribute
	Read AttributeAccess = 1 << iota
	Write
	SetByCreate
)

func (access AttributeAccess) String() string {
	switch access {
	default:
		return "Unknown"
	case Read:
		return "Read"
	case Write:
		return "Write"
	case SetByCreate:
		return "SetByCreate"
	case Read | Write:
		return "Read/Write"
	case Read | SetByCreate:
		return "Read/SetByCreate"
	case Write | SetByCreate:
		return "Write/SetByCreate"
	case Read | Write | SetByCreate:
		return "Read/Write/SetByCreate"
	}
}

// SupportsAttributeAccess returns true if the managed entity attribute
// supports the desired access
func SupportsAttributeAccess(attr AttributeDefinition, acc AttributeAccess) bool {
	return attr.GetAccess().Contains(acc)
}

type IManagedEntityDefinition interface {
	GetName() string
	GetClassID() ClassID
	GetMessageTypes() mapset.Set
	GetAllowedAttributeMask() uint16
	GetAttributeDefinitions() AttributeDefinitionMap
	GetAlarmMap() AlarmMap

	DecodeAttributes(uint16, []byte, gopacket.PacketBuilder, byte) (AttributeValueMap, error)
	SerializeAttributes(AttributeValueMap, uint16, gopacket.SerializeBuffer, byte, int, bool) (error, uint16)
}

type IManagedEntity interface {
	IManagedEntityDefinition
	GetManagedEntityDefinition() IManagedEntityDefinition

	GetEntityID() uint16
	SetEntityID(uint16) error

	GetAttributeMask() uint16

	GetAttributeValueMap() AttributeValueMap
	GetAttribute(string) (interface{}, error)
	GetAttributeByIndex(uint) (interface{}, error)

	SetAttribute(string, interface{}) error
	SetAttributeByIndex(uint, interface{}) error

	DeleteAttribute(string) error
	DeleteAttributeByIndex(uint) error
}