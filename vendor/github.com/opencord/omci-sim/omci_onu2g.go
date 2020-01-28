/*
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

package core

import (
	"encoding/binary"
)

type Onu2GAttributes int

const (
	_                                           = iota
	EquipmentID                 Onu2GAttributes = 0x8000
	OmccVersion                 Onu2GAttributes = 0x4000
	VendorProductCode           Onu2GAttributes = 0x2000
	SecurityCapability          Onu2GAttributes = 0x1000
	SecurityMode                Onu2GAttributes = 0x0800
	TotalPriorityQueueNumber    Onu2GAttributes = 0x0400
	TotalTrafficSchedulerNumber Onu2GAttributes = 0x0200
	Mode                        Onu2GAttributes = 0x0100
	TotalGemPortIDNumber        Onu2GAttributes = 0x0080
	SysUptime                   Onu2GAttributes = 0x0040
	ConnectivityCapability      Onu2GAttributes = 0x0020
	CurrentConnectivityMode     Onu2GAttributes = 0x0010
	QosConfigurationFlexibility Onu2GAttributes = 0x0008
	PriorityQueueScaleFactor    Onu2GAttributes = 0x0004
)

type Onu2GAttributeHandler func(*uint, []byte) ([]byte, error)

var Onu2GAttributeHandlers = map[Onu2GAttributes]Onu2GAttributeHandler{
	EquipmentID:                 GetEquipmentID,
	OmccVersion:                 GetOmccVersion,
	VendorProductCode:           GetVendorProductCode,
	SecurityCapability:          GetSecurityCapability,
	SecurityMode:                GetSecurityMode,
	TotalPriorityQueueNumber:    GetTotalPriorityQueueNumber,
	TotalTrafficSchedulerNumber: GetTotalTrafficSchedulerNumber,
	Mode:                        GetMode,
	TotalGemPortIDNumber:        GetTotalGemPortIDNumber,
	SysUptime:                   GetSysUptime,
	ConnectivityCapability:      GetConnectivityCapability,
	CurrentConnectivityMode:     GetCurrentConnectivityMode,
	QosConfigurationFlexibility: GetQosConfigurationFlexibility,
	PriorityQueueScaleFactor:    GetPriorityQueueScaleFactor,
}

func GetOnu2GAttributes(pos *uint, pkt []byte, content OmciContent) ([]byte, error) {
	AttributesMask := getAttributeMask(content)

	for index := uint(16); index >= 1; index-- {
		Attribute := 1 << (index - 1)
		reqAttribute := Attribute & AttributesMask

		if reqAttribute != 0 {
			pkt, _ = Onu2GAttributeHandlers[Onu2GAttributes(reqAttribute)](pos, pkt)
		}
	}

	pkt[8] = 0x00 // Command Processed Successfully
	pkt[9] = uint8(AttributesMask >> 8)
	pkt[10] = uint8(AttributesMask & 0x00FF)

	return pkt, nil

}

func GetEquipmentID(pos *uint, pkt []byte) ([]byte, error) {
	// 20 bytes
	equipid := []byte("12345123451234512345")
	for _, ch := range equipid {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetOmccVersion(pos *uint, pkt []byte) ([]byte, error) {
	// 1 bytes
	pkt[*pos] = 0xB4
	*pos++
	return pkt, nil
}

func GetVendorProductCode(pos *uint, pkt []byte) ([]byte, error) {
	// 2 bytes
	prodcode := []byte{0x00, 0x00}
	for _, ch := range prodcode {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetSecurityCapability(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetSecurityMode(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetTotalPriorityQueueNumber(pos *uint, pkt []byte) ([]byte, error) {
	// 2 bytes
	// report 0 queues because thats what BRCM does...
	numqueues := 0
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(numqueues))
	for _, ch := range bs {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetTotalTrafficSchedulerNumber(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetMode(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetTotalGemPortIDNumber(pos *uint, pkt []byte) ([]byte, error) {
	// 2 bytes
	gemports := 32
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(gemports))
	for _, ch := range bs {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetSysUptime(pos *uint, pkt []byte) ([]byte, error) {
	// 4 byte int
	uptime := 0
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(uptime))
	for _, ch := range bs {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetConnectivityCapability(pos *uint, pkt []byte) ([]byte, error) {
	// 2 bytes
	caps := []byte{0x00, 0x7F}
	for _, ch := range caps {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetCurrentConnectivityMode(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetQosConfigurationFlexibility(pos *uint, pkt []byte) ([]byte, error) {
	// 2 bytes
	qosconf := []byte{0x00, 0x30}
	for _, ch := range qosconf {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetPriorityQueueScaleFactor(pos *uint, pkt []byte) ([]byte, error) {
	// 1 bytes
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}
